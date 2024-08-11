// Copyright (c) 2024 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blacknon/go-sshlib"
	"github.com/blacknon/lssh/output"
	"golang.org/x/crypto/ssh"
)

var (
	pShellHelptext = `{{.Name}} - {{.Usage}}

	{{.HelpName}} {{if .VisibleFlags}}[options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}
	{{range .VisibleFlags}}	{{.}}
	{{end}}
	`
)

// TODO(blacknon): 以下のBuild-in Commandを追加する
//     - %cd <PATH>         ... リモートのディレクトリを変更する(事前のチェックにsftpを使用か？)
//     - %lcd <PATH>        ... ローカルのディレクトリを変更する
//     - %save <num> <PATH> ... 指定したnumの履歴をPATHに記録する (v0.6.11)
//     - %set <args..>      ... 指定されたオプションを設定する(Optionsにて管理) (v0.6.11)
//     - %diff <num>        ... 指定されたnumの履歴をdiffする(multi diff)。できるかどうか要検討。 (v0.7.0以降)
//                              できれば、vimdiffのように横に差分表示させるようにしたいものだけど…？
//     - %get remote local  ... sftpプロトコルを利用して、ファイルやディレクトリを取得する (v0.6.11)
//     - %put local remote  ... sftpプロトコルを利用して、ファイルやディレクトリを配置する (v0.6.11)

// TODO(blacknon): 任意のBuild-in Commandを追加できるようにする
//    - configにて、環境変数に過去のoutの出力をつけて任意のスクリプトを実行できるようにしてやることで、任意のスクリプト実行が可能に出来たら良くないか？というネタ
//    - もしくは、Goのモジュールとして機能追加できるようにするって方法もありかも？？

// checkBuildInCommand return true if cmd is build-in command.
func checkBuildInCommand(cmd string) (isBuildInCmd bool) {
	// check build-in command
	switch cmd {
	case "exit", "quit", "clear": // build-in command
		isBuildInCmd = true

	case
		"%history",
		"%out", "%outlist", "%outexec",
		"%save",
		"%set": // parsent build-in command.
		isBuildInCmd = true
	}

	return
}

// checkLocalCommand return bool, check is pshell build-in command or
// local machine command(%%command).
func checkLocalCommand(cmd string) (isLocalCmd bool) {
	// check local command regex
	regex := regexp.MustCompile(`^!.*`)

	// local command
	switch {
	case regex.MatchString(cmd):
		isLocalCmd = true
	}

	return
}

// check local or build-in command
func checkLocalBuildInCommand(cmd string) (result bool) {
	// check build-in command
	result = checkBuildInCommand(cmd)
	if result {
		return result
	}

	// check local command
	result = checkLocalCommand(cmd)

	return result
}

// runBuildInCommand is run buildin or local machine command.
func (s *shell) run(pline pipeLine, in *io.PipeReader, out *io.PipeWriter, ch chan<- bool, kill chan bool) (err error) {
	// get 1st element
	command := pline.Args[0]

	// check and exec build-in command
	switch command {
	// exit or quit
	case "exit", "quit":
		os.Exit(0)

	// clear
	case "clear":
		fmt.Printf("\033[H\033[2J")
		return

	// %history
	case "%history":
		s.buildin_history(out, ch)
		return

	// %outlist
	case "%outlist":
		s.buildin_outlist(out, ch)
		return

	// %out [num]
	case "%out":
		num := s.Count - 1
		if len(pline.Args) > 1 {
			num, err = strconv.Atoi(pline.Args[1])
			if err != nil {
				return
			}
		}

		s.buildin_out(num, out, ch)
		return

	// %outexec [num]
	case "%outexec":
		s.buildin_outexec(pline, in, out, ch, kill)
		return
	}

	// check and exec local command
	buildinRegex := regexp.MustCompile(`^!.*`)
	switch {
	case buildinRegex.MatchString(command):
		// exec local machine
		s.executeLocalPipeLine(pline, in, out, ch, kill, os.Environ())
	default:
		// exec remote machine
		s.executeRemotePipeLine(pline, in, out, ch, kill)
	}

	return
}

// localCmd_set is set pshll option.
// TODO(blacknon): Optionsの値などについて、あとから変更できるようにする。
// func (s *shell) buildin_set(args []string, out *io.PipeWriter, ch chan<- bool) {
// }

// localCmd_save is save HistoryResult results as a file local.
//     %save num PATH(独自の環境変数を利用して個別のファイルに保存できるようにする)
// TODO(blacknon): Optionsの値などについて、あとから変更できるようにする。
// func (s *shell) buildin_save(args []string, out *io.PipeWriter, ch chan<- bool) {
// }

// localCmd_history is printout history (shell history)
func (s *shell) buildin_history(out *io.PipeWriter, ch chan<- bool) {
	stdout := setOutput(out)

	// read history file
	data, err := s.GetHistoryFromFile()
	if err != nil {
		return
	}

	// print out history
	for _, h := range data {
		fmt.Fprintf(stdout, "%s: %s\n", h.Timestamp, h.Command)
	}

	// close out
	switch stdout.(type) {
	case *io.PipeWriter:
		out.CloseWithError(io.ErrClosedPipe)
	}

	// send exit
	ch <- true
}

// localcmd_outlist is print exec history list.
func (s *shell) buildin_outlist(out *io.PipeWriter, ch chan<- bool) {
	stdout := setOutput(out)

	for i := 0; i < len(s.History); i++ {
		h := s.History[i]
		for _, hh := range h {
			fmt.Fprintf(stdout, "%3d : %s\n", i, hh.Command)
			break
		}
	}

	// close out
	switch stdout.(type) {
	case *io.PipeWriter:
		out.CloseWithError(io.ErrClosedPipe)
	}

	// send exit
	ch <- true
}

// localCmd_out is print exec history at number
// example:
//   - %out
//   - %out <num>
func (s *shell) buildin_out(num int, out *io.PipeWriter, ch chan<- bool) {
	stdout := setOutput(out)
	histories := s.History[num]

	// get key
	keys := []string{}
	for k := range histories {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	i := 0
	for _, k := range keys {
		h := histories[k]

		// if first, print out command
		if i == 0 {
			fmt.Fprintf(os.Stderr, "[History:%s ]\n", h.Command)
		}
		i += 1

		// print out result
		if len(histories) > 1 && stdout == os.Stdout && h.Output != nil {
			// set Output.Count
			bc := h.Output.Count
			h.Output.Count = num
			op := h.Output.GetPrompt()

			// TODO(blacknon): Outputを利用させてOPROMPTを生成
			sc := bufio.NewScanner(strings.NewReader(h.Result))
			for sc.Scan() {
				fmt.Fprintf(stdout, "%s %s\n", op, sc.Text())
			}

			// reset Output.Count
			h.Output.Count = bc
		} else {
			fmt.Fprintf(stdout, h.Result)
		}
	}

	// close out
	switch stdout.(type) {
	case *io.PipeWriter:
		out.CloseWithError(io.ErrClosedPipe)
	}

	// send exit
	ch <- true
}

// executePipeLineRemote is exec command in remote machine.
// Didn't know how to send data from Writer to Channel, so switch the function if * io.PipeWriter is Nil.
func (s *shell) executeRemotePipeLine(pline pipeLine, in *io.PipeReader, out *io.PipeWriter, ch chan<- bool, kill chan bool) {
	// join command
	command := strings.Join(pline.Args, " ")

	// set stdin/stdout
	stdin := setInput(in)
	stdout := setOutput(out)

	// create channels
	exit := make(chan bool)
	exitInput := make(chan bool) // Input finish channel
	exitOutput := make(chan bool)

	// create []io.WriteCloser
	var writers []io.WriteCloser

	// create []ssh.Session
	var sessions []*ssh.Session

	// create session and writers
	m := new(sync.Mutex)
	for _, c := range s.Connects {
		// create session
		session, err := c.CreateSession()
		if err != nil {
			continue
		}

		// Request tty (Only when input is os.Stdin and output is os.Stdout).
		if stdin == os.Stdin && stdout == os.Stdout {
			sshlib.RequestTty(session)
		}

		// set stdout
		var ow io.Writer
		ow = stdout
		if ow == os.Stdout {
			// create Output Writer
			c.Output.Count = s.Count
			w := c.Output.NewWriter()
			defer w.CloseWithError(io.ErrClosedPipe)

			// create pShellHistory Writer
			hw := s.NewHistoryWriter(c.Output.Server, c.Output, m)
			defer hw.CloseWithError(io.ErrClosedPipe)

			ow = io.MultiWriter(w, hw)
		}
		session.Stdout = ow

		// get and append stdin writer
		w, _ := session.StdinPipe()
		writers = append(writers, w)

		// append sessions
		sessions = append(sessions, session)
	}

	// multi input-writer
	go output.PushInput(exitInput, writers, stdin)

	// run command
	for _, s := range sessions {
		session := s
		go func() {
			session.Run(command)
			session.Close()
			exit <- true
			if stdout == os.Stdout {
				exitOutput <- true
			}
		}()
	}

	// kill
	go func() {
		select {
		case <-kill:
			for _, s := range sessions {
				s.Signal(ssh.SIGINT)
				s.Close()
			}
		}
	}()

	// wait
	s.wait(len(sessions), exit)

	// wait time (0.050 sec)
	time.Sleep(500 * time.Millisecond)

	// send exit
	ch <- true

	// exit input.
	if stdin == os.Stdin {
		exitInput <- true
	}

	// close out
	switch stdout.(type) {
	case *io.PipeWriter:
		out.CloseWithError(io.ErrClosedPipe)
	}

	// wait time (0.050 sec)
	time.Sleep(500 * time.Millisecond)

	return
}

// executePipeLineLocal is exec command in local machine.
// TODO(blacknon): 利用中のShellでの実行+functionや環境変数、aliasの引き継ぎを行えるように実装
func (s *shell) executeLocalPipeLine(pline pipeLine, in *io.PipeReader, out *io.PipeWriter, ch chan<- bool, kill chan bool, envrionment []string) (err error) {
	// set stdin/stdout
	stdin := setInput(in)
	stdout := setOutput(out)

	// set HistoryResult
	var stdoutw io.Writer
	stdoutw = stdout
	m := new(sync.Mutex)
	if stdout == os.Stdout {
		pw := s.NewHistoryWriter("localhost", nil, m)
		defer pw.CloseWithError(io.ErrClosedPipe)
		stdoutw = io.MultiWriter(pw, stdout)
	} else {
		stdoutw = stdout
	}

	// delete command prefix(`!`)
	rep := regexp.MustCompile(`^!`)
	pline.Args[0] = rep.ReplaceAllString(pline.Args[0], "")

	// join command
	command := strings.Join(pline.Args, " ")

	// execute command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// set stdin, stdout, stderr
	cmd.Stdin = stdin
	if s.Options.LocalCommandNotRecordResult {
		cmd.Stdout = stdout
	} else { // default
		cmd.Stdout = stdoutw
	}
	cmd.Stderr = os.Stderr

	// set envrionment
	cmd.Env = envrionment

	// run command
	err = cmd.Start()

	// get signal and kill
	p := cmd.Process
	go func() {
		select {
		case <-kill:
			p.Kill()
		}
	}()

	// wait command
	cmd.Wait()

	// close out, or write pShellHistory
	switch stdout.(type) {
	case *io.PipeWriter:
		out.CloseWithError(io.ErrClosedPipe)
	}

	// send exit
	ch <- true

	return
}

// s.wait
func (s *shell) wait(num int, ch <-chan bool) {
	for i := 0; i < num; i++ {
		<-ch
	}
}

// setInput
func setInput(in io.ReadCloser) (stdin io.ReadCloser) {
	if reflect.ValueOf(in).IsNil() {
		stdin = os.Stdin
	} else {
		stdin = in
	}

	return
}

// setOutput
func setOutput(out io.WriteCloser) (stdout io.WriteCloser) {
	if reflect.ValueOf(out).IsNil() {
		stdout = os.Stdout
	} else {
		stdout = out
	}

	return
}
