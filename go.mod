module github.com/blacknon/lsshell

require (
	github.com/blacknon/go-sshlib v0.1.12
	github.com/blacknon/lssh v0.6.9
	github.com/c-bata/go-prompt v0.2.6
	github.com/urfave/cli v1.22.15
	golang.org/x/crypto v0.24.0
	mvdan.cc/sh v2.6.4+incompatible
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/ScaleFT/sshkeys v0.0.0-20200327173127-6142f742bca5 // indirect
	github.com/ThalesIgnite/crypto11 v1.2.5 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5 // indirect
	github.com/blacknon/go-x11auth v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/dchest/bcrypt_pbkdf v0.0.0-20150205184540-83f37f9c154a // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190724205821-6cfae18c12b8 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/nsf/termbox-go v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sevlyar/go-daemon v0.1.5 // indirect
	github.com/thales-e-security/pool v0.0.2 // indirect
	github.com/vbauerster/mpb v3.4.0+incompatible // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.21.0 // indirect
)

// replace
replace (
	github.com/ThalesIgnite/crypto11 v1.2.5 => github.com/blacknon/crypto11 v1.2.6
	github.com/blacknon/lssh v0.6.9 => ../lssh
	github.com/c-bata/go-prompt v0.2.6 => github.com/blacknon/go-prompt v0.2.7
)

go 1.22.4

toolchain go1.22.5
