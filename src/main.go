package main

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Pom struct {
	XMLName    xml.Name `xml:"project"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Packaging  string   `xml:"packaging"`
	Parent     Parent   `xml:"parent"`
}

type Parent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type DeployFile struct {
	FilePath  string
	PomConfig Pom
}

const (
	RootPath       string = "/Users/xxxx/developer/ot_jar/leatop/leatop-msdp-parent"
	RemoteURL      string = "https://nexus.xxx.cn/repository/leatop-snapshot/"
	RerepositoryId string = "leatop-snapshot"
)

// 读取pom 文件，获取坐标
func getPom(path string) (Pom, error) {
	fmt.Printf("path: %v\n", path)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return Pom{}, err
	}
	// 关闭流
	defer file.Close()
	// 读取xml 文件
	b, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		log.Fatal(err2)
		return Pom{}, err2
	}
	s := strings.Split(path, "/")
	// 取上一层目录,作为版本
	version := s[len(s)-2 : len(s)-1]
	fmt.Printf("version: %v\n", version)

	v := Pom{}
	err3 := xml.Unmarshal(b, &v)
	if err != nil {
		log.Fatal(err3)
	}
	if v.GroupId == "" {
		v.GroupId = v.Parent.GroupId
	}
	if v.Packaging == "" {
		v.Packaging = "jar"
	}
	v.Version = version[0]
	// fmt.Printf("v: %v\n", v)
	return v, nil
}

// 查找需要发布的jar文件
func findDeployFile() ([]DeployFile, error) {

	var deployeFilesSlice []DeployFile

	err := filepath.Walk(RootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if !info.IsDir() {
			s := strings.Split(path, "/")
			// 取上一层目录,作为版本
			version := s[len(s)-2 : len(s)-1]
			if strings.Contains(info.Name(), version[0]) && strings.HasSuffix(info.Name(), ".pom") {
				p, _ := getPom(path)
				if strings.Contains(info.Name(), p.Version) {
					var _path = path
					if p.Packaging == "jar" {
						_path = strings.Replace(_path, ".pom", ".jar", 1)
					}
					var depolyeFile = DeployFile{FilePath: _path, PomConfig: p}
					deployeFilesSlice = append(deployeFilesSlice, depolyeFile)
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deployeFilesSlice, nil
}

// 发布到私服
func deployeCMD(deployConfig DeployFile) {
	cmd := exec.Command("mvn",
		"deploy:deploy-file",
		"-Dmaven.test.skip=true",
		"-Dfile="+deployConfig.FilePath,
		"-DgroupId="+deployConfig.PomConfig.GroupId,
		"-DartifactId="+deployConfig.PomConfig.ArtifactId,
		"-Dversion="+deployConfig.PomConfig.Version,
		"-Dpackaging="+deployConfig.PomConfig.Packaging,
		"-DrepositoryId="+RerepositoryId,
		"-Durl="+RemoteURL)

	fmt.Printf("cmd.Args: %v\n", cmd.Args)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// 关闭流
	defer stdout.Close()
	// 输出命令执行结果
	if err := cmd.Start(); err != nil { // 运行命令
		log.Fatal(err)
	}
	if opBytes, err := ioutil.ReadAll(stdout); err != nil { // 读取输出结果
		log.Fatal(err)
		return
	} else {
		log.Println(string(opBytes))
	}
}

func main() {

	df, _ := findDeployFile()

	for _, file := range df {
		fmt.Printf("file: %v\n", file)
		deployeCMD(file)
	}

	// cmd := exec.Command("mvn", "-version")

}
