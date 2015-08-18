package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func resource() *schema.Resource {
	return &schema.Resource{
		Create: Create,
		Delete: Delete,
		Exists: Exists,
		Read:   Read,

		Schema: map[string]*schema.Schema{
			"command": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "command to run",
				ForceNew:    true,
				// Make a "best effort" attempt to relativize the file path.
				StateFunc: func(v interface{}) string {
					pwd, err := os.Getwd()
					if err != nil {
						return v.(string)
					}
					rel, err := filepath.Rel(pwd, v.(string))
					if err != nil {
						return v.(string)
					}
					return rel
				},
			},
			"json_output": &schema.Schema{
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "command output as json. nil if unparsable.",
			},
			"text_output": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "command output as string",
			},
		},
	}
}

func Create(d *schema.ResourceData, meta interface{}) error {
	command := d.Get("command").(string)
	output, json_output := run_cmd(command)
	d.Set("text_output", output)
	d.Set("json_output", json_output)
	d.SetId(hash(command+output))
	return nil
}

func Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func Exists(d *schema.ResourceData, meta interface{}) (bool, error) {
	command := d.Get("command").(string)
	output, _ := run_cmd(command)
	return hash(command+output) == d.Id(), nil
}

func Read(d *schema.ResourceData, meta interface{}) error {
	// Logic is handled in Exists, which only returns true if the rendered
	// contents haven't changed. That means if we get here there's nothing to
	// do.
	return nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func run_cmd(s string) (string, map[string]interface{}) {
	ss := safeSplit(s)
	cmdName := ss[0]
	cmdArgs := ss[1:]
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Command `%v` failed with `%v`\n", s, err)
		return "", nil
	}
	// attempt json decode
	var m map[string]interface{}
	err = json.Unmarshal(cmdOut, &m)
	if err != nil {
		log.Println(err)
		m = nil
	}
	return string(cmdOut), m
}

// stealing this function for cmdline from
// <https://gist.github.com/jmervine/d88c75329f98e09f5c87>
func safeSplit(s string) []string {
	split := strings.Split(s, " ")

	var result []string
	var inquote string
	var block string
	for _, i := range split {
		if inquote == "" {
			if strings.HasPrefix(i, "'") || strings.HasPrefix(i, "\"") {
				inquote = string(i[0])
				block = strings.TrimPrefix(i, inquote) + " "
			} else {
				result = append(result, i)
			}
		} else {
			if !strings.HasSuffix(i, inquote) {
				block += i + " "
			} else {
				block += strings.TrimSuffix(i, inquote)
				inquote = ""
				result = append(result, block)
				block = ""
			}
		}
	}

	return result
}
