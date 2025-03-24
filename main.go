package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/pkg/browser"
	"github.com/pkg/errors"

	"github.com/feight/deploy/schema"
	"github.com/feight/deploy/tui"
)

type DeployTarget interface {
	Configure(serviceName string, env []string)
	Deploy()
	GetEnvironment() []string
	GetImageRegistry() string
	GetImageTag() string
	GetProject() string
	PostDeploy()
	SetKey(string)
	Text() string
}

var (
	conf *Config
)

func main() {
	defer onError()

	saveSchema()

	env := getEnv()
	conf = env.defaultConfig()

	/*
	 * Select environment.
	 */

	if len(env.Config) > 1 {
		conf = tui.RenderList(env.Config, "e", "Which environment would you like to use?")
	}

	fmt.Printf("\n⏺ %s: %s\n", color.WhiteString("Environment"), conf.SelectedText())

	/*
	 * Select service to deploy.
	 */

	service := tui.RenderList(conf.Services, "s", "What would you like to deploy?")
	fmt.Printf("\n⏺ %s: %s\n", color.WhiteString("Deployment"), color.CyanString(service.SelectedText()))

	/*
	 * Select deployment target.
	 */

	target := service.defaultTarget()

	if len(service.targetMap()) > 1 {
		target = tui.RenderList(service.targetMap(), "t", "Where would you like to deploy?")
	}

	fmt.Printf("\n⏺ %s: %s", color.WhiteString("Deploy target"), color.CyanString(target.Text()))

	/*
	 * start.
	 */

	fmt.Println()
	target.Configure(service.key, conf.GlobalEnv)
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
	t.Deploy()

	took := time.Since(start).Round(time.Millisecond * 100).String()

	fmt.Printf("\n🎉 Successfully deployed %s to %s in %s.\n\n", color.YellowString(s.Name), color.YellowString(t.GetProject()), took)

	beeep.Alert("Deployment", fmt.Sprintf("🎉 Successfully deployed %s to %s in %s.", s.Name, t.GetProject(), took), "")

	postDeploy(s, t)

	t.PostDeploy()

	cleanUp()

	if s.Open != "" {
		browser.OpenURL(s.Open)
	}
}

func createImage(s *Service, t DeployTarget) {

	runPrebuild(s, t)
	runBuild(s, t)

	fmt.Printf("\n> Creating %s docker image [%s]...\n\n", color.YellowString(s.Name), t.GetImageTag())
	runBuildImage(s, t)

	fmt.Printf("\n> Pushing %s image to container registry...\n\n", color.YellowString(s.Name))
	pushImage(s, t)
}

func runPrebuild(s *Service, t DeployTarget) {

	if len(s.Prebuild) < 1 {
		return
	}

	name, args := bash(s.Prebuild)

	err := command(s, t, name, args...).Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func runBuild(s *Service, t DeployTarget) {

	if conf.UseTurboRepo {

		runTurbo(s, t, "clean")
		runTurbo(s, t, "build", "--env-mode=loose", "--force")
	} else {

		err := command(s, t, "npm", "run", "build").Run()

		if err != nil {
			panic(errors.Wrap(err, "could not run npm build"))
		}
	}
}

func runTurbo(s *Service, t DeployTarget, args ...string) {

	if _, err := os.Stat(path.Join(s.Path, "package.json")); err != nil {
		return
	}

	cmd := command(s, t, "npx", "turbo")
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(cmd.Env, []string{"TURBO_NO_UPDATE_NOTIFIER=true"}...)

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not run turbo"))
	}
}

func postDeploy(s *Service, t DeployTarget) {

	if len(s.Postdeploy) < 1 {
		return
	}

	name, args := bash(s.Postdeploy)

	err := command(s, t, name, args...).Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func runBuildImage(s *Service, t DeployTarget) {

	cmd := command(
		s,
		t,
		"docker",
		"build",
		"--platform", "linux/amd64",
		"-t", t.GetImageTag(),
	)

	if s.Dockerfile != "" {
		cmd.Args = append(cmd.Args, "--file", s.Dockerfile)
	}

	cmd.Args = append(cmd.Args, ".")

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete build"))
	}
}

func pushImage(s *Service, t DeployTarget) {

	cmd := command(
		s,
		t,
		"docker",
		"push",
		t.GetImageTag(),
	)

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

func command(s *Service, t DeployTarget, name string, arg ...string) *exec.Cmd {

	cmd := exec.Command(name, arg...)

	cmd.Dir = s.Path
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, t.GetEnvironment()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func bash(command string) (string, []string) {

	return "bash", []string{"-c", command}
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

func (s *Env) defaultConfig() *Config {
	for _, c := range s.Config {
		return c
	}
	return nil
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

func (s *Config) Text() string {

	if s.Name != "" {
		return fmt.Sprintf("%-20s (%s)", s.Name, s.key)
	}
	return s.key
}

func (s *Config) SelectedText() string {

	ret := ""
	if s.Name != "" {
		ret = color.CyanString("%s (%s)", s.Name, s.key)
	} else {
		ret = color.CyanString(s.key)
	}
	if s.IsProduction {
		ret += color.HiYellowString("  ⚠️ production")
	}

	return ret
}

func (s *Config) SetKey(key string) {
	s.key = key
}

func (s *Service) Text() string {

	padding := len(s.Name)
	for _, service := range conf.Services {
		if len(service.Name) > padding {
			padding = len(service.Name)
		}
	}

	if len(s.targetMap()) > 1 {
		return s.Name + " ..."
	} else {
		return fmt.Sprintf("%-*s %s", padding+5, s.Name, s.defaultTarget().Text())
	}
}

func (s *Service) SelectedText() string {
	return s.Name
}

func (s *Service) SetKey(key string) {
	s.key = key
}

func (s *Service) targetMap() map[string]DeployTarget {

	ret := map[string]DeployTarget{}
	interfaceType := reflect.TypeOf((*DeployTarget)(nil)).Elem()
	v := reflect.ValueOf(*s.Targets)
	for i := range v.NumField() {
		f := v.Field(i)
		typ := v.Type().Field(i)
		if typ.Type.Implements(interfaceType) && typ.IsExported() && !f.IsNil() {
			ret[strconv.Itoa(i)] = f.Interface().(DeployTarget)
		}
	}

	return ret
}

func (s *Service) defaultTarget() DeployTarget {
	for _, t := range s.targetMap() {
		return t
	}
	return nil
}

func saveSchema() {

	do := flag.Bool("writeSchema", false, "write the json schema for configuration")
	flag.Parse()

	if *do {
		schema, _ := schema.GetSchema(Config{})
		b, _ := json.MarshalIndent(schema, "", "    ")
		os.WriteFile("schema.json", b, os.ModePerm)
	}
}
