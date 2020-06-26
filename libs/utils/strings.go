package utils

import (
	"bytes"
	"encoding/base64"
	"github.com/hashicorp/go-version"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

func Reverse(strs []string) []string {
	outs := make([]string, len(strs))
	for i := 0; i < len(outs)/2; i++ {
		outs[i], outs[len(outs)-i-1] = strs[len(outs)-i-1], strs[i]
	}
	if len(outs)%2 == 1 {
		outs[len(outs)/2] = strs[len(outs)/2]
	}
	return outs
}

func Base64File(file string) string {
	bs, err := ioutil.ReadFile(file)
	Panic(err, "read file %s", file)
	return base64.StdEncoding.EncodeToString(bs)
}

func NotExists(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}

func Exists(path string) bool {
	return !NotExists(path)
}

func FileBytes(path string) []byte {
	bs, err := ioutil.ReadFile(path)
	Panic(err, "read file %s", path)
	return bs
}

func DashName(name string) string {
	return Name(name, "-")
}

func UnderscoreName(name string) string {
	return Name(name, "_")
}

func Name(name, link string) string {
	buffer := bytes.NewBufferString("")
	preis := false
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 && !preis {
				buffer.WriteString(link)
			}
			preis = true
			buffer.WriteRune(unicode.ToLower(r))
		} else {
			preis = false
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

// 下划线写法转为驼峰写法
func CamelName(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Replace(name, "-", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

func Search(slice []string, s string) int {
	for i, v := range slice {
		if s == v {
			return i
		}
	}
	return -1
}

func FirstLine(content string) string {
	return strings.Split(content, "\n")[0]
}

func Match(str string, patterns ...string) bool {
	for _, pattern := range patterns {
		if regexp.MustCompile(pattern).MatchString(str) {
			return true
		}
	}
	return false
}

func SelectMatch(strs []string, patterns ...string) (matched []string) {
	matched = make([]string, 0)
	for _, str := range strs {
		if Match(str, patterns...) {
			matched = append(matched, str)
		}
	}
	return matched
}

func SelectNotMatch(strs []string, patterns ...string) (matched []string) {
	matched = make([]string, 0)
	for _, str := range strs {
		if !Match(str, patterns...) {
			matched = append(matched, str)
		}
	}
	return matched
}

func Join(ms map[string]string, separator, kvSeparator string) string {
	out := bytes.NewBufferString("")
	i := 0
	for k, v := range ms {
		if i != 0 {
			out.WriteString(separator)
		}
		i++
		out.WriteString(k)
		out.WriteString(kvSeparator)
		out.WriteString(v)
	}
	return out.String()
}
func KeyJoin(ms map[string]string, separator, kvSeparator string) string {
	out := bytes.NewBufferString("")
	i := 0
	for k, _ := range ms {
		if i != 0 {
			out.WriteString(separator)
		}
		i++
		out.WriteString(k)
		out.WriteString(kvSeparator)
	}
	return out.String()
}

func Repeat(str string, size int) []string {
	outs := make([]string, size)
	for i := 0; i < size; i++ {
		outs[i] = str
	}
	return outs
}

func Append(ips []string, app string) []string {
	outs := make([]string, 0)
	for _, ip := range ips {
		outs = append(outs, ip+""+app)
	}
	return outs
}

//版本比较
func VersionCompose(v1, v2 string) int {
	v11, err := version.NewVersion(v1)
	Panic(err, v1)
	v12, err := version.NewVersion(v2)
	Panic(err, v2)
	return v11.Compare(v12)
}

func Split2(str, sep string) (a, b string) {
	outs := strings.SplitN(str, sep, 2)
	a = outs[0]
	if len(outs) == 2 {
		b = outs[1]
	}
	return
}

func Split3(str, sep string) (a, b, c string) {
	outs := strings.SplitN(str, sep, 3)
	a = outs[0]
	if len(outs) > 1 {
		b = outs[1]
	}
	if len(outs) > 2 {
		c = outs[2]
	}
	return
}

func Switch(assert bool, a, b string) string {
	if assert {
		return a
	} else {
		return b
	}
}

func Index(args []string, index int) string {
	if index > len(args)-1 {
		return ""
	}
	return args[index]
}
