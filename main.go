package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	flag "github.com/spf13/pflag"
)

//go:embed data/*
var fs embed.FS

//go:embed dataDyn/*
var fsDyn embed.FS

// todo better naming

func main() {

	id := flag.String("id", "com.system76.CosmicAppletExample", "App ID")
	icon := flag.String("icon", "display-symbolic", "Icon name")
	name := flag.StringP("name", "n", "cosmic-applet-example", "App name")
	icon_files := flag.StringSlice("icon-files", []string{}, "path to icon files")
	interactive_ := flag.BoolP("interactive", "i", false, "Activate interactive mode")
	flag.Parse()
	if *interactive_ {
		interactive(id, icon, name, icon_files)
	}
	fmt.Printf("Your input, are you sure? id: %v, icon: %v, name: %v, icon_files: %v \n", *id, *icon, *name, *icon_files)
	exit := ""
	must(fmt.Scan(&exit))
	mustBool(strings.HasPrefix(strings.ToLower(exit), "y"))

	a := newApp(*id, *icon, *name, *icon_files)
	a.write()
	fmt.Printf("\nDone - Now Type:\n\ncd %v \ncargo b -r \nsudo just install\n", a.Name)

}

type app struct {
	t *template.Template

	// file names
	f              []string
	ID, Icon, Name string
}

func (a *app) write() {
	for _, n := range a.f {
		var p string
		if strings.HasSuffix(n, ".rs") {
			p = filepath.Join(a.Name, "src", n)
		} else if strings.HasSuffix(n, ".desktop") {
			p = filepath.Join(a.Name, "data", n)
		} else {
			p = filepath.Join(a.Name, n)
		}

		f := must(os.Create(p))
		defer f.Close()
		must1(a.t.ExecuteTemplate(f, n, a))
	}

}
func (a *app) IconName() string {
	return strings.TrimSuffix(a.Icon, filepath.Ext(a.Icon))
}
func (a *app) FormatName() string {
	return strings.ReplaceAll(a.Name, "-", " ")
}
func newApp(id, icon, name string, icon_files []string) *app {

	f := must(fs.ReadDir("data"))
	f1 := make([]string, 0, len(f)+1)
	for i := range f {
		f1 = append(f1, f[i].Name())
	}
	must1(os.MkdirAll(filepath.Join(name, "data"), os.ModePerm))
	f1 = append(f1, fmt.Sprintf("%v.desktop", id))
	t := must(template.New("").Funcs(template.FuncMap{
		"replace": strings.ReplaceAll,
	}).ParseFS(fs, "data/*"))
	f10 := must(fsDyn.Open("dataDyn/id.desktop"))
	defer f10.Close()
	must(t.New(f1[len(f1)-1]).Parse(string(must(io.ReadAll(f10)))))

	addDep := "cargo add --git 'https://github.com/pop-os/libcosmic' libcosmic --no-default-features -F 'libcosmic/applet,libcosmic/tokio,libcosmic/wayland'"
	addDep2 := "cargo add rust-embed i18n-embed-fl once_cell i18n-embed -F 'i18n-embed/fluent-system,i18n-embed/desktop-requester'"
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %v && cargo init && %v && %v", name, addDep, addDep2))
	fmt.Println("Executing command:", cmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	must1(cmd.Run())
	p := filepath.Join(name, "i18n", "en")
	must1(os.MkdirAll(p, os.ModePerm))
	// todo touch

	f2 := must(os.Create(filepath.Join(p, fmt.Sprintf("%v.ftl", strings.ReplaceAll(name, "-", "_")))))
	f2.Close()
	f3 := must(os.Create(filepath.Join(name, "data", fmt.Sprintf("%v.desktop", id))))
	f3.Close()
	appsDir := filepath.Join(name, "data", "icons", "scalable", "apps")
	must1(os.MkdirAll(appsDir, os.ModePerm))
	for _, p := range icon_files {
		must1(os.WriteFile(filepath.Join(appsDir, filepath.Base(p)), must(os.ReadFile(p)), os.ModePerm))
	}
	return &app{
		t:    t,
		ID:   id,
		f:    f1,
		Icon: icon,
		Name: name,
	}
}

func interactive(id, icon, name *string, icon_files *[]string) {
	fmt.Println("Interacitve Mode:")
	prompts := map[string]*string{"id": id, "icon": icon, "name": name}
	s := bufio.NewScanner(os.Stdin)
	for k, v := range prompts {
		fmt.Printf("%v: ", k)
		mustBool(s.Scan())
		t := s.Text()
		if t == "" {
			continue
		}
		*v = t
	}
	fmt.Printf("icon files: ")
	mustBool(s.Scan())
	t := s.Text()
	if t != "" {
		*icon_files = strings.Fields(t)
	}
	must1(s.Err())

}

func mustBool(b bool) {
	if !b {
		panic("false")
	}
}

func must1(e error) {
	if e != nil {
		panic(e)
	}
}
func must[T any](t T, e error) T {
	if e != nil {
		panic(e)
	}
	return t
}
