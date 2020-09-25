package main

import (
	"fmt"
	"github.com/shoriwe/FullProxy/src/ArgumentParser"
	"github.com/shoriwe/FullProxy/src/MasterSlave"
	"github.com/shoriwe/FullProxy/src/Proxies/PortForward"
	"github.com/shoriwe/FullProxy/src/Proxies/SOCKS5"
	"github.com/shoriwe/FullProxy/src/Proxies/Translation/ForwardToSocks5"
	"os"
)


func main() {
	numberOfArguments := len(os.Args)
	if numberOfArguments == 1 {
		_, _ = fmt.Fprintln(os.Stderr, "Try:\n", os.Args[0], " help")
		os.Exit(-1)
	}
	switch os.Args[1] {
	case "socks5":
		address, port, username, password, slave := ArgumentParser.ParseSocks5Arguments()
		SOCKS5.StartSocks5(address, port, slave, &username, &password)
	case "http":
		fmt.Println("HTTP not implemented yet")
	case "local-forward":
		address, port, masterAddress, masterPort := ArgumentParser.LocalPortForwardArguments()
		PortForward.StartLocalPortForward(address, port, masterAddress, masterPort)
	case "remote-forward":
		localAddress, localPort, masterAddress, masterPort := ArgumentParser.RemotePortForwardArguments()
		PortForward.StartRemotePortForward(localAddress, localPort, masterAddress, masterPort)
	case "master":
		masterAddress, masterPort, remoteAddress, remotePort := ArgumentParser.ParseMasterArguments()
		MasterSlave.Master(masterAddress, masterPort, remoteAddress, remotePort)
	case "translate":
		if numberOfArguments == 2 {
			_, _ = fmt.Fprintln(os.Stderr, "Try:\n", os.Args[0], " translate help")
			os.Exit(-1)
		}
		switch os.Args[2] {
		case "forward-socks5":
			bindAddress, bindPort, socks5Address, socks5Port, username, password, targetAddress, targetPort := ArgumentParser.ParseForwardToSocks5Arguments()
			ForwardToSocks5.StartForwardToSocks5Translation(bindAddress, bindPort, socks5Address, socks5Port, username, password, targetAddress, targetPort)
		case "help":
			ArgumentParser.ShowTranslateHelpMessage()
		default:
			_, _ = fmt.Fprintln(os.Stderr, "Unknown command\nTry: ", os.Args[0], " translate help")
			os.Exit(-1)
		}
	case "help":
		ArgumentParser.ShowGeneralHelpMessage()
	default:
		_, _ = fmt.Fprintln(os.Stderr, "Unknown command\nTry: ", os.Args[0], " help")
		os.Exit(-1)
	}
}
