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
}

type DeployFile struct {
	FilePath  string
	PomConfig Pom
}

const (
	RootPath       string = "/Users/herongqin/.m2/repository/com/zerosky/framework/zerosky-framework-util"
	RemoteURL      string = "https://nexus.zerosky.cn/repository/maven-releases/"
	RerepositoryId string = "maven-releases"
)

// 读取pom 文件，获取坐标
func getPom(path string) (Pom, error) {
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
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jar") {
			// fmt.Printf("info.Name(): %v\n", info.Name())
			p, _ := getPom(strings.Replace(path, "jar", "pom", 1))
			var depolyeFile = DeployFile{FilePath: path, PomConfig: p}
			deployeFilesSlice = append(deployeFilesSlice, depolyeFile)

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
		"-Dpackaging=jar",
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
