package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	target := "/Users/lianghao/PhpStorm/js.dtm.cc"
	task := fmt.Sprintf("扫描目录%s，并读取目录下的 .txt 文件，读取第一个txt文件的前五行")

	fmt.Println("任务是：", task)
	fmt.Println()

	fmt.Println("===== 普通LLM回复 ======")
	fmt.Println(chatBotReply(target))

	fmt.Println()
	fmt.Println("===== Agent回复 ======")
	fmt.Println(agentReply(target))
}

func chatBotReply(target string) string {
	return `1. xxxx
2. xxxx
3. xxxx`
}

func agentReply(target string) string {
	var (
		out   strings.Builder
		files []string
	)

	_ = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".txt") {
			files = append(files, path)
		}
		return nil
	})

	out.WriteString("\n第一步：扫描目录 得到.json文件" + strings.Join(files, "\n"))
	if len(files) == 0 {
		out.WriteString("没有文件")
		return out.String()
	}

	first := files[0]
	out.WriteString("\n第二步：读取" + first + "前五行\n")
	lines, err := readFirstFiveLines(first, 5)
	if err != nil {
		out.WriteString("读取失败" + err.Error())
	}
	//out.WriteString(strings.Join(lines, "\n"))

	out.WriteString("\n第三步：整理结果，first：" + first + ", \n结果\n" + strings.Join(lines, "\n"))
	return out.String()
}

func readFirstFiveLines(file string, n int) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) >= n {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
