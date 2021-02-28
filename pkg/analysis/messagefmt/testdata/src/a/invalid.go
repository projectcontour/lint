package a

import (
	"github.com/sirupsen/logrus"
)

func invalid() {
	logrus.Info("A walrus appears")    // want `message starts with uppercase: "A walrus appears"`
	logrus.Debug("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Error("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Fatal("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Panic("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Print("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Info("A walrus appears")    // want `message starts with uppercase: "A walrus appears"`
	logrus.Trace("A walrus appears")   // want `message starts with uppercase: "A walrus appears"`
	logrus.Warn("A walrus appears")    // want `message starts with uppercase: "A walrus appears"`
	logrus.Warning("A walrus appears") // want `message starts with uppercase: "A walrus appears"`
}
