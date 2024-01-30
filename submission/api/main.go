package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/jtav77/Janelle/submission/api/internal/hooks"
	_ "github.com/jtav77/Janelle/submission/api/internal/migrations"
)

func main() {
	ctx := context.TODO()
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangGo,
		Automigrate:  false,
	})

	partTag := os.Getenv("PARTICIPANT_TAG")
	if partTag == "" {
		panic("required env var not set: PARTICIPANT_TAG")
	}

	dockerRepo := os.Getenv("DOCKER_REPO")
	if dockerRepo == "" {
		panic("required env var not set: DOCKER_REPO")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	hooks.AddHooks(ctx, app, clientset)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
