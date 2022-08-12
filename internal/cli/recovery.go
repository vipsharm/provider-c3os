package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	config "github.com/c3os-io/c3os/pkg/config"
	"github.com/ipfs/go-log"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/mudler/edgevpn/pkg/logger"
	"github.com/mudler/edgevpn/pkg/node"
	"github.com/mudler/edgevpn/pkg/services"
	"github.com/pterm/pterm"
)

func startRecoveryService(ctx context.Context, token, name, address, loglevel string) error {

	nc := config.Network(token, "", loglevel, "c3osrecovery0")

	lvl, err := log.LevelFromString(loglevel)
	if err != nil {
		lvl = log.LevelError
	}
	llger := logger.New(lvl)

	o, _, err := nc.ToOpts(llger)
	if err != nil {
		llger.Fatal(err.Error())
	}

	o = append(o,
		services.Alive(
			time.Duration(20)*time.Second,
			time.Duration(10)*time.Second,
			time.Duration(10)*time.Second)...)

	// opts, err := vpn.Register(vpnOpts...)
	// if err != nil {
	// 	return err
	// }
	o = append(o, services.RegisterService(llger, time.Duration(5*time.Second), name, address)...)

	e, err := node.New(o...)
	if err != nil {
		return err
	}

	return e.Start(ctx)
}

func sshServer(listenAdddr, password string) {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("bash")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				pterm.Warning.Println("Failed reserving tty")
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()
			go func() {
				io.Copy(f, s) //nolint:errcheck
			}()
			io.Copy(s, f) //nolint:errcheck
			cmd.Wait()    //nolint:errcheck
		} else {
			io.WriteString(s, "No PTY requested.\n") //nolint:errcheck
			s.Exit(1)                                //nolint:errcheck
		}
	})

	pterm.Info.Println(ssh.ListenAndServe(listenAdddr, nil, ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
		return pass == password
	}),
	))
}

func StartRecoveryService(tk, serviceUUID, generatedPassword, listenAddr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := startRecoveryService(ctx, tk, serviceUUID, listenAddr, "fatal"); err != nil {
		return err
	}

	sshServer(listenAddr, generatedPassword)

	return fmt.Errorf("should not return")
}