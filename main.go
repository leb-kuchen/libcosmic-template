package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	flag "github.com/spf13/pflag"
)

//go:embed data/*
var fs embed.FS

//go:embed dataDyn/*
var fsDyn embed.FS

//go:embed version
var versionString string

// todo add tests for config
// todo better naming
// todo go-flags
// todo name should be template
// conditional filenames, parse walkfn

func main() {
	id := flag.String("id", "com.system76.CosmicAppletExample", "App ID")
	icon := flag.String("icon", "display-symbolic", "Icon name")
	name := flag.StringP("name", "n", "cosmic-applet-example", "App name")
	icon_files := flag.StringSlice("icon-files", []string{}, "path to icon files")
	interactive_ := flag.BoolP("interactive", "i", false, "Activate interactive mode")
	config := flag.BoolP("config", "c", true, "Generate cosmic-config")
	noConfirm := flag.Bool("no-confirm", false, "Do not ask for confirmation")
	noExample := flag.Bool("no-example", false, "Do not generate example row")
	version := flag.BoolP("version", "V", false, "Show the latest version")
	flag.Parse()
	if *version {
		fmt.Println(versionString)
		return
	}
	if *interactive_ {
		interactive(id, icon, name, icon_files)
	}

	client := &http.Client{}
	urlWg := &sync.WaitGroup{}
	// todo file based
	crates := []string{
		"rust-embed",
		"i18n-embed-fl",
		"once_cell",
		"i18n-embed",
		"serde",
		"pop-os/libcosmic github.com",
	}
	versions := make([]string, len(crates))
	for idx, crate := range crates {
		crateName, _, ok := strings.Cut(crate, " ")
		var versioner Versioner
		var url string
		if !ok {
			urlWg.Add(1)
			url = fmt.Sprintf("https://crates.io/api/v1/crates/%v/versions", crateName)
			versioner = &CrateInfo{}
			go fetch(versioner, url, crate, idx, client, urlWg, versions)

		} //else {
		// 	versioner = &GithubInfo{}
		// 	url = fmt.Sprintf("https://api.%v/repos/%v/tags", provider, crateName)
		// }

		// todo get latest rev
	}
	urlWg.Wait()
	if !*noConfirm {
		fmt.Printf("\nYour input, are you sure?\n\nid: %v\nicon: %v\nname: %v\nicon_files: %v\n\n", *id, *icon, *name, *icon_files)
		exit := ""
		must(fmt.Scan(&exit))
		mustBool(strings.HasPrefix(strings.ToLower(exit), "y"))
		// todo next time run ...

	}
	a := newApp(*id, *icon, *name, *config, !*noExample, *icon_files, versions)
	a.write()
	fmt.Printf("\nDone - Now Type:\n\ncd %v \ncargo b -r \nsudo just install\n", a.Name)

}

type app struct {
	t *template.Template

	Versions []string
	// file names
	f               []string
	ID, Icon, Name  string
	Config, Example bool
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

// todo test recursive folders
// todo only one template folder

func newApp(id, icon, name string, config, example bool, icon_files, versions []string) *app {
	f := must(fs.ReadDir("data"))
	fDynOpen := must(fsDyn.ReadDir("dataDyn"))
	f1 := make([]string, 0, len(f)+len(fDynOpen))
	for i := range f {
		f1 = append(f1, f[i].Name())
	}
	must1(os.MkdirAll(filepath.Join(name, "data"), os.ModePerm))
	must1(os.MkdirAll(filepath.Join(name, "src"), os.ModePerm))

	t := must(template.New("").Funcs(template.FuncMap{
		"replace": strings.ReplaceAll,
	}).ParseFS(fs, "data/*"))
	// obviously function
	// embed fs is readonly
	// todo own fs?
	for i := range fDynOpen {
		var buf strings.Builder
		filename := fDynOpen[i].Name()
		fileTemplate := must(template.New("").Parse(filename))
		fileTemplate.Execute(&buf, map[string]any{"ID": id, "Config": config})
		parsedFilename := strings.TrimSpace(buf.String())
		if parsedFilename == "" {
			continue
		}
		fileContent := must(fsDyn.ReadFile(filepath.Join("dataDyn", filename)))
		f1 = append(f1, parsedFilename)
		must(t.New(parsedFilename).Parse(string(fileContent)))
	}

	// for now if ftl i18n/en
	p := filepath.Join(name, "i18n", "en")
	must1(os.MkdirAll(p, os.ModePerm))
	// todo touch
	// todo move logic to templates folder
	// todo rename data to templates

	// could be write file
	//  {{.Name}}.ftl
	f2 := must(os.Create(filepath.Join(p, fmt.Sprintf("%v.ftl", strings.ReplaceAll(name, "-", "_")))))
	f2.Close()
	appsDir := filepath.Join(name, "data", "icons", "scalable", "apps")
	must1(os.MkdirAll(appsDir, os.ModePerm))
	// copy icon files
	// could be at the end
	for _, p := range icon_files {
		must1(os.WriteFile(filepath.Join(appsDir, filepath.Base(p)), must(os.ReadFile(p)), os.ModePerm))
	}
	return &app{
		t:        t,
		ID:       id,
		f:        f1,
		Icon:     icon,
		Name:     name,
		Versions: versions,
		Config:   config,
		Example:  example,
	}
}

// func parseFS() {

// }

// todo propbably should be interface
func fetch(content Versioner, url, crate string, idx int, client *http.Client, urlWg *sync.WaitGroup, versions []string) {
	defer urlWg.Done()
	fmt.Println("Fetching:", url)
	req := must(http.NewRequest("GET", url, nil))
	// strings.Trim?
	userAgent := fmt.Sprintf("libcosmic-template/%s github.com/leb-kuchen/libcomsic-template", versionString)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	res := must(client.Do(req))
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic(res.Status)
	}
	body := must(io.ReadAll(res.Body))
	must1(json.Unmarshal(body, &content))
	version := content.version()
	versions[idx] = version
	fmt.Printf("%v: %v\n", crate, version)
}

type CrateInfo struct {
	Versions []struct {
		Num string `json:"num"`
	} `json:"versions"`
}
type GithubInfo []struct {
	Name string `json:"name"`
}

func (a CrateInfo) version() string {
	return a.Versions[0].Num
}

func (a GithubInfo) version() string {
	return a[0].Name
}

type Versioner interface {
	version() string
}

func interactive(id, icon, name *string, icon_files *[]string) {
	// todo map is random
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
