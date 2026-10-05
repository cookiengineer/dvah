package tools

import "dvah/schemas"
import utils_fmt "dvah/utils/fmt"
import utils_fs "dvah/utils/fs"
import "context"
import "fmt"
import "os"
import "os/exec"
import "slices"
import "sort"
import "strings"
import "time"

type Files struct {
	Methods []string
	Sandbox string
}

func NewFiles(methods []string, sandbox string) *Files {

	return &Files{
		Methods: methods,
		Sandbox: sandbox,
	}

}

func (tool *Files) Name() string {
	return "files"
}

func (tool *Files) Call(method string, arguments map[string]interface{}) (string, error) {

	if tool.HasMethod(method) == true {

		if method == "Copy" {

			from_path, ok1 := arguments["from_path"].(string)
			to_path,   ok2 := arguments["to_path"].(string)

			if ok1 == true && ok2 == true {
				return tool.Copy(utils_fmt.FormatFilePath(from_path), utils_fmt.FormatFilePath(to_path))
			} else if ok1 == true && ok2 == false {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"to_path\" is not a string.")
			} else if ok1 == false && ok2 == true {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"from_path\" is not a string.")
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameters.")
			}

		} else if method == "List" {

			path, ok := arguments["path"].(string)

			if ok == true {
				return tool.List(utils_fmt.FormatFilePath(path))
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"path\" is not a string.")
			}

		} else if method == "Read" {

			path, ok := arguments["path"].(string)

			if ok == true {
				return tool.Read(utils_fmt.FormatFilePath(path))
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"path\" is not a string.")
			}

		} else if method == "Search" {

			keywords, ok := arguments["keywords"].(string)

			if ok == true {
				return tool.Search(keywords)
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"keywords\" is not a string.")
			}

		} else if method == "Stat" {

			path, ok := arguments["path"].(string)

			if ok == true {
				return tool.Stat(utils_fmt.FormatFilePath(path))
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"path\" is not a string.")
			}

		} else if method == "Write" {

			path,    ok1 := arguments["path"].(string)
			content, ok2 := arguments["content"].(string)

			if ok1 == true && ok2 == true {
				return tool.Write(utils_fmt.FormatFilePath(path), content)
			} else if ok1 == true && ok2 == false {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"content\" is not a string.")
			} else if ok1 == false && ok2 == true {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameter \"path\" is not a string.")
			} else {
				return "", fmt.Errorf("files.%s: %s", method, "Invalid parameters.")
			}

		} else {
			return "", fmt.Errorf("files.%s: Invalid method.", method)
		}

	} else {
		return "", fmt.Errorf("files.%s: Method not allowed.", method)
	}

}

func (tool *Files) Copy(from_path string, to_path string) (string, error) {

	from_resolved, err1 := resolveSandboxPath(tool.Sandbox, from_path)

	if err1 == nil {

		to_resolved, err2 := resolveSandboxPath(tool.Sandbox, to_path)

		if err2 == nil {

			stat, err3 := os.Stat(from_resolved)

			if err3 == nil {

				if stat.IsDir() {

					err4 := utils_fs.CopyAll(from_resolved, to_resolved)

					if err4 == nil {
						return fmt.Sprintf("files.Copy: Folder \"%s\" copied to \"%s\".", from_path, to_path), nil
					} else {
						return "", fmt.Errorf("files.Copy: Cannot copy \"%s\" to \"%s\".", from_path, to_path)
					}

				} else {

					err4 := utils_fs.Copy(from_resolved, to_resolved)

					if err4 == nil {
						return fmt.Sprintf("files.Copy: File \"%s\" copied to \"%s\".", from_path, to_path), nil
					} else {
						return "", fmt.Errorf("files.Copy: Cannot copy \"%s\" to \"%s\".", from_path, to_path)
					}

				}

			} else {
				return sanitizeFilesystemError("files", "Copy", "Path", from_path, err3)
			}

		} else {
			return "", fmt.Errorf("files.Copy: %s", err2.Error())
		}

	} else {
		return "", fmt.Errorf("files.Copy: %s", err1.Error())
	}

}

func (tool *Files) GetContent(id string) (any, error) {
	return nil, nil
}

func (tool *Files) GetContentIdentifiers() []string {
	return []string{}
}

func (tool *Files) HasMethod(method string) bool {
	return slices.Contains(tool.Methods, method) == true
}

func (tool *Files) List(path string) (string, error) {

	if path == "/" {
		path = "."
	} else if path == "" {
		path = "."
	}

	resolved, err0 := resolveSandboxPath(tool.Sandbox, path)

	if err0 == nil {

		stat, err1 := os.Stat(resolved)

		if err1 == nil {

			if stat.IsDir() == true {

				entries, err2 := os.ReadDir(resolved)

				if err2 == nil {

					lines := make([]string, 0)

					for _, entry := range entries {

						name      := entry.Name()
						path, err := sanitizeSandboxPath(tool.Sandbox, resolved + "/" + name)

						if err == nil && strings.HasPrefix(name, ".") == false {

							typ := "file"

							if entry.IsDir() == true {
								typ = "folder"
							}

							lines = append(lines, fmt.Sprintf("- Path: \"%s\", Type: %s", path, typ))

						}

					}

					sort.Strings(lines)

					result := make([]string, 0)
					result = append(result, fmt.Sprintf("files.List: \"%s\" contains %d entries.", path, len(lines)))

					for l := 0; l < len(lines); l++ {
						result = append(result, lines[l])
					}

					return strings.Join(result, "\n"), nil

				} else {
					return "", fmt.Errorf("files.List: Cannot list folder \"%s\".", path)
				}

			} else {
				return "", fmt.Errorf("files.List: Invalid folder path \"%s\".", path)
			}

		} else {
			return sanitizeFilesystemError("files", "List", "Folder", path, err1)
		}

	} else {
		return "", fmt.Errorf("files.List: %s", err0.Error())
	}

}

func (tool *Files) Read(path string) (string, error) {

	resolved, err0 := resolveSandboxPath(tool.Sandbox, path)

	if err0 == nil {

		bytes, err1 := os.ReadFile(resolved)

		if err1 == nil {

			result := strings.Join([]string{
				fmt.Sprintf("files.Read: File \"%s\" contents", path),
				string(bytes),
			}, "\n")

			return result, nil

		} else {
			return sanitizeFilesystemError("files", "Read", "File", path, err1)
		}

	} else {
		return "", fmt.Errorf("files.Read: %s", err0.Error())
	}

}

// Search recursively greps the sandbox for the given keywords.
//
// NOTE: This is intentionally vulnerable. The keywords are interpolated into a
// shell command verbatim, so shell metacharacters (;, |, &, $(), backticks,
// newlines, ...) are interpreted by /bin/sh. This allows arbitrary command
// execution as well as Shellshock-style payloads smuggled through the keywords.
func (tool *Files) Search(keywords string) (string, error) {

	command := fmt.Sprintf("grep -rn %s .", keywords)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = tool.Sandbox

	cmd.Stdin = strings.NewReader("")

	output, err0 := cmd.CombinedOutput()

	result := strings.Join([]string{
		fmt.Sprintf("files.Search: grep -rn %s .", keywords),
		strings.TrimSpace(string(output)),
	}, "\n")

	if err0 != nil {

		// grep exits with code 1 when it simply found no matches.
		if exit_err, ok := err0.(*exec.ExitError); ok == true && exit_err.ExitCode() == 1 {
			return result, nil
		}

		return result, fmt.Errorf("files.Search: %s", err0.Error())

	}

	return result, nil

}

func (tool *Files) Schemas() []schemas.Tool {

	result := make([]schemas.Tool, 0)

	for _, method := range tool.Methods {

		for _, schema := range FilesSchema {

			if schema.Function.Name == fmt.Sprintf("%s.%s", tool.Name(), method) {
				result = append(result, schema)
			}

		}

	}

	return result

}

func (tool *Files) Stat(path string) (string, error) {

	resolved, err0 := resolveSandboxPath(tool.Sandbox, path)

	if err0 == nil {

		stat, err1 := os.Stat(resolved)

		if err1 == nil {

			typ := "file"

			if stat.IsDir() == true {
				typ = "folder"
			}

			result := strings.Join([]string{
				fmt.Sprintf("files.Stat: \"%s\" is a %s.", path, typ),
				"Name: " + stat.Name(),
				"Type: " + typ,
				"Size: " + utils_fmt.FormatFileSize(stat.Size()),
				"Mode: " + utils_fmt.FormatFileMode(stat.Mode()),
				"Modified: " + utils_fmt.FormatTime(stat.ModTime()),
			}, "\n")

			return result, nil

		} else {
			return sanitizeFilesystemError("files", "Stat", "File", path, err1)
		}

	} else {
		return "", fmt.Errorf("files.Stat: %s", err0.Error())
	}

}

func (tool *Files) Write(path string, content string) (string, error) {

	resolved, err0 := resolveSandboxPath(tool.Sandbox, path)

	if err0 == nil {

		buffer, err1 := utils_fmt.FormatFileBuffer(content)

		if err1 == nil {

			err2 := os.WriteFile(resolved, buffer, 0666)

			if err2 == nil {

				result := strings.Join([]string{
					fmt.Sprintf("files.Write: File \"%s\" with %s written.", path, utils_fmt.FormatFileSize(int64(len(buffer)))),
				}, "\n")

				return result, nil

			} else {
				return sanitizeFilesystemError("files", "Write", "Folder", path, err2)
			}

		} else {
			return "", fmt.Errorf("files.Write: %s", err1.Error())
		}

	} else {
		return "", fmt.Errorf("files.Write: %s", err0.Error())
	}

}
