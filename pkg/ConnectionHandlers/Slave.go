package ConnectionHandlers

import (
	"crypto/tls"
	"github.com/shoriwe/FullProxy/pkg/Sockets"
	"log"
	"net"
	"time"
)

func ConnectToMasterServer(masterAddress *string, masterPort *string) (net.Conn, *tls.Config) {
	tlsConfiguration, configurationCreationError := Sockets.CreateSlaveTLSConfiguration()
	if configurationCreationError == nil {
		log.Printf("Trying to connecto to %s:%s", *masterAddress, *masterPort)
		masterConnection, connectionError := Sockets.TLSConnect(masterAddress, masterPort, tlsConfiguration)
		log.Printf("Successfully connected to %s:%s", *masterAddress, *masterPort)
		if connectionError == nil {
			return masterConnection, tlsConfiguration
		}
		log.Fatal("Could not connect to the master")
	}
	log.Fatal("Could not create the TLS certificate")
	return nil, nil
}

func GeneralSlave(masterAddress *string, masterPort *string, protocol ProxyProtocol) {
	masterConnection, tlsConfiguration := ConnectToMasterServer(masterAddress, masterPort)
	masterConnectionReader, _ := Sockets.CreateSocketConnectionReaderWriter(masterConnection)
	for {
		_ = masterConnection.SetReadDeadline(time.Now().Add(20 * time.Second))
		NumberOfReceivedBytes, buffer, connectionError := Sockets.Receive(masterConnectionReader, 1024)
		if connectionError == nil {
			if NumberOfReceivedBytes == 1 {
				if buffer[0] == NewConnection[0] {
					clientConnection, connectionError := Sockets.TLSConnect(masterAddress, masterPort, tlsConfiguration)

					if connectionError == nil {
						clientConnectionReader, clientConnectionWriter := Sockets.CreateSocketConnectionReaderWriter(clientConnection)
						if clientConnectionReader != nil && clientConnectionWriter != nil {
							go protocol.Handle(clientConnection, clientConnectionReader, clientConnectionWriter)
						} else {
							_ = clientConnection.Close()
						}
					}
				}
			} else {
				continue
			}
		} else if parsedConnectionError, ok := connectionError.(net.Error); !(ok && parsedConnectionError.Timeout()) {
			log.Print(connectionError)
			break
		}
	}
}

func RemotePortForwardSlave(
	masterAddress *string, masterPort *string,
	localAddress *string, localPort *string,
	protocol ProxyProtocol) {
	localServer := Sockets.Bind(localAddress, localPort)
	masterConnection, tlsConfiguration := ConnectToMasterServer(masterAddress, masterPort)
	masterConnectionReader, masterConnectionWriter := Sockets.CreateSocketConnectionReaderWriter(masterConnection)
	for {
		clientConnection, connectionError := localServer.Accept()
		if connectionError == nil {
			log.Print("Connection from: ", clientConnection.RemoteAddr().String())
			_, connectionError := Sockets.Send(masterConnectionWriter, &NewConnection)
			if connectionError == nil {
				_ = masterConnection.SetReadDeadline(time.Now().Add(3 * time.Second))
				numberOfBytesReceived, response, connectionError := Sockets.Receive(masterConnectionReader, 1)
				if connectionError == nil {
					if numberOfBytesReceived == 1 {
						switch response[0] {
						case NewConnection[0]:
							targetConnection, connectionError := Sockets.TLSConnect(masterAddress, masterPort, tlsConfiguration)
							if connectionError == nil {
								go startGeneralProxying(clientConnection, targetConnection)
							} else {
								_ = clientConnection.Close()
								log.Fatal("Connectivity issues with master server")
							}
						case FailToConnectToTarget[0]:
							_ = clientConnection.Close()
							log.Print("Something goes wrong when master connected to target")
							break
						case UnknownOperation[0]:
							_ = clientConnection.Close()
							log.Print("The master did not understood the message")
							break
						}
					}
				} else if parsedConnectionError, ok := connectionError.(net.Error); !(ok && parsedConnectionError.Timeout()) {
					log.Print(connectionError)
					break
				}
			}
		}
	}
}
