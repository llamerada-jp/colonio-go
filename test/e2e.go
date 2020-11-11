package test

import (
	"os"
	"os/exec"

	colonio "github.com/llamerada-jp/colonio-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testE2E() {
	It("does E2E test", func() {
		By("starting seed for test")
		cur, _ := os.Getwd()
		cmd := exec.Command(os.Getenv("COLONIO_SEED_BIN_PATH"), "-config", cur+"/seed_config.json")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		Expect(err).ShouldNot(HaveOccurred())
		defer func() {
			err = cmd.Process.Kill()
			Expect(err).ShouldNot(HaveOccurred())
		}()

		By("creating a new colonio instance")
		c, err := colonio.NewColonio()
		Expect(err).ShouldNot(HaveOccurred())

		By("connecting to seed")
		Eventually(func() error {
			return c.Connect("ws://localhost:8080/test", "")
		}).ShouldNot(HaveOccurred())

		By("disconnecting from seed")
		err = c.Disconnect()
		Expect(err).ShouldNot(HaveOccurred())

		By("quiting colonio instance")
		err = c.Quit()
		Expect(err).ShouldNot(HaveOccurred())
	})
}
