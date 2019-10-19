package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type commit struct {
	Author  string `xml:"author"`
	Project string `xml:"project"`
	Date    string `xml:"date"`
	Message string `xml:"message"`
}

func info(app *cli.App) {
	app.Name = "Daily Standup Helper CLI"
	app.Usage = "Reports git history"
	app.Author = "github.com/Lebonesco"
	app.Version = "1.0.0"
}

func commands(app *cli.App) {
	app.Action = func(c *cli.Context) error {
		dir := c.String("dir")
		after := c.String("after")

		user := c.String("user")
		if len(user) == 0 {
			return fmt.Errorf("no 'user' flag value provided")
		}

		commits, err := getGitHistory(dir, user, after)
		if err != nil {
			return err
		}

		f, err := os.Create("standup.json")
		if err != nil {
			return err
		}

		prettyJSON, err := json.MarshalIndent(commits, "", "  ")
		if err != nil {
			return err
		}

		_, err = f.Write(prettyJSON)
		if err != nil {
			return err
		}

		log.Println("completed...")
		return nil
	}
}

func flags(app *cli.App) {
	dir, _ := homedir.Dir()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Value: "",
			Usage: "git user name",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: dir,
			Usage: "parent directory to start recursively searching for *.git files",
		},
		cli.StringFlag{
			Name:  "after, a",
			Value: time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05"),
			Usage: "when to start looking at commit history",
		},
	}
}

func getGitHistory(dir, user, after string) ([]commit, error) {
	var commits []commit
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".git" {
			b, err := getCommits(path, user, after)
			if err != nil {
				return err
			}

			if len(b) == 0 {
				return nil
			}

			// https://stackoverflow.com/questions/27553274/unmarshal-xml-array-in-golang-only-getting-the-first-element
			//https://yourbasic.org/golang/list-files-in-directory/
			d := xml.NewDecoder(bytes.NewBuffer(b))
			for {
				var c commit
				err := d.Decode(&c)
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}

				c.Project = getParentDir(path)
				commits = append(commits, c)
			}

		}

		return nil
	})

	return commits, err
}

func getParentDir(path string) string {
	ss := strings.Split(path, "/")
	if len(ss) < 2 {
		return ""
	}
	parentDir := ss[len(ss)-2]
	return parentDir
}

func getCommits(path, user, after string) ([]byte, error) {
	format := `
			<entry>
				<author>%an</author>
				<date>%cd</date>
				<message>%B</message>
			</entry>`
	cmd := exec.Command("git", "log", "--author="+user, "--pretty=format:"+format, "--after="+after)
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// https://itnext.io/how-to-create-your-own-cli-with-golang-3c50727ac608
func main() {
	app := cli.NewApp()
	info(app)
	flags(app)
	commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
