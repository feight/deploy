package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/pkg/browser"
	"github.com/pkg/errors"

	"github.com/feight/deploy/tui"
)

type DeployTarget interface {
	Text() string
	GetProject() string
	GetImageRegistry() string
	GetImageTag(*Service) string
	Deploy(*Service)
	PostDeploy(*Service)
}

var (
	conf *Config
)

func Start() {
	defer onError()

	env := getEnv()

	for _, conf = range env.Config {
		break
	}

	/*
	 * Select environment.
	 */

	if len(env.Config) > 1 {
		conf = tui.RenderList(env.Config, "s", "Which environment would you like to use?")
		fmt.Printf("\n⏺ %s: %s\n", color.WhiteString("Environment"), color.CyanString(conf.key))
	}

	/*
	 * Select service to deploy.
	 */

	service := tui.RenderList(conf.Services, "s", "What would you like to deploy?")
	fmt.Printf("\n⏺ %s: %s\n", color.WhiteString("Deployment"), color.CyanString(service.Name))

	/*
	 * Select deployment target.
	 */

	target := tui.RenderList(service.Targets, "t", "Where would you like to deploy?").get()
	fmt.Printf("\n⏺ %s: %s", color.WhiteString("Deploy target"), color.CyanString(target.Text()))

	/*
	 * start.
	 */

	fmt.Println()
	start(service, target)
}

func onError() {
	e := recover()

	if e != nil {
		fmt.Println(color.HiRedString("\n🔥 %s\n", e))
	}
}

func start(s *Service, t DeployTarget) {

	var (
		start = time.Now()
	)

	fmt.Println()

	if s.Path != "" {
		createImage(s, t)
	}

	fmt.Printf("\n> Deploying %s to %s...\n\n", color.YellowString(s.Name), color.YellowString(t.GetProject()))
	t.Deploy(s)

	took := time.Since(start).Round(time.Millisecond * 100).String()

	fmt.Printf("\n🎉 Successfully deployed %s to %s in %s.\n\n", color.YellowString(s.Name), color.YellowString(t.GetProject()), took)

	beeep.Alert("Deployment", fmt.Sprintf("🎉 Successfully deployed %s to %s in %s.", s.Name, t.GetProject(), took), "")

	postDeploy(s)

	t.PostDeploy(s)

	cleanUp()

	if s.Open != "" {
		browser.OpenURL(s.Open)
	}
}

func createImage(s *Service, t DeployTarget) {

	runPrebuild(s)
	runBuild(s)

	fmt.Printf("\n> Creating %s docker image...\n\n", color.YellowString(s.Name))
	runBuildImage(s, t)

	fmt.Printf("\n> Pushing %s image to container registry...\n\n", color.YellowString(s.Name))
	pushImage(s, t)
}

func runPrebuild(s *Service) {

	if len(s.Prebuild) < 1 {
		return
	}

	name, args := bash(s.Prebuild)

	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Dir = s.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func runBuild(s *Service) {

	if conf.UseTurboRepo {

		runTurbo(s, "clean")
		runTurbo(s, "build",
			"--output-logs=errors-only",
			"--no-cache",
			"--force")
	} else {

		cmd := exec.Command("npm", "run", "build")
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()

		if err != nil {
			panic(errors.Wrap(err, "could not run npm build"))
		}
	}
}

func runTurbo(s *Service, args ...string) {

	if _, err := os.Stat(path.Join(s.Path, "package.json")); err != nil {
		return
	}

	cmd := exec.Command("npx", "turbo")

	cmd.Args = append(cmd.Args, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{"TURBO_NO_UPDATE_NOTIFIER=true"}...)
	cmd.Dir = s.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not run turbo"))
	}
}

func postDeploy(s *Service) {

	if len(s.Postdeploy) < 1 {
		return
	}

	name, args := bash(s.Postdeploy)

	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Dir = s.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func runBuildImage(s *Service, t DeployTarget) {

	dockerfilePath := s.Path

	if s.Dockerfile != "" {
		dockerfilePath = path.Join(dockerfilePath, s.Dockerfile)
	}

	cmd := exec.Command(
		"docker",
		"build",
		"--platform", "linux/amd64",
		"-t", t.GetImageTag(s),
		dockerfilePath,
	)

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func pushImage(s *Service, t DeployTarget) {

	cmd := exec.Command(
		"docker",
		"push",
		t.GetImageTag(s),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrapf(err,
			`could not push image. you may need to run "gcloud auth configure-docker %s"`, t.GetImageRegistry()))
	}
}

func cleanUp() {

	cmd := exec.Command(
		"docker",
		"image",
		"prune",
		"-f",
	)

	cmd.Stderr = os.Stderr
	cmd.Run()
}

func bash(command string) (string, []string) {

	return "bash", []string{"-c", command}
}

func (s *Target) get() DeployTarget {

	if s.Cloudrun != nil {
		return s.Cloudrun
	}
	if s.Kube != nil {
		return s.Kube
	}
	if s.Registry != nil {
		return s.Registry
	}
	if s.CloudLoadBalancer != nil {
		return s.CloudLoadBalancer
	}

	panic("invalid target")
}

func getEnv() Env {
	env := Env{
		Config: map[string]*Config{},
	}

	matches, err := filepath.Glob("deploy*.json")
	if err != nil {
		panic(errors.Wrap(err, "could not get files"))
	}

	for _, filename := range matches {
		env.Config[filename] = getConfig(filename)
	}

	return env
}

func getConfig(filename string) *Config {

	bin, err := os.ReadFile(filename)
	if err != nil {
		panic(errors.Wrapf(err, "could not open %s", filename))
	}

	conf := Config{
		key: filename,
	}

	json.Unmarshal(bin, &conf)

	return &conf
}

// Combine global environment variables with e
func env(e []string) []string {
	return append(conf.GlobalEnv, e...)
}

func (s *Config) Text() string {
	return s.key
}

func (s *Config) SetKey(key string) {
	s.key = key
}

func (s *Service) Text() string {
	return s.Name
}

func (s *Service) SetKey(key string) {
	s.key = key
}

func (s *Target) Text() string {
	return s.get().Text()
}

func (s *Target) SetKey(key string) {}
