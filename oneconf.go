/*
Package oneconf will populate a central configuration struct wil data from TOML files, default, environment and command line.
*/
package oneconf

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
)

func scan(v any, chain []string, cb func(name, kind string, tag reflect.StructTag, chain []string) string) {
	rt := reflect.TypeOf(v).Elem()
	rv := reflect.ValueOf(v).Elem()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i).Name
		// vt := rv.Field(i).Type()
		vv := rv.Field(i).Addr()

		switch rv.Field(i).Kind() {
		case reflect.String:
			// fmt.Printf("%s : %s(%s)-%v\n", field, vv, vt, tag)
			if val := cb(field, "string", rt.Field(i).Tag, chain); val != "" {
				rv.Field(i).SetString(val)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// fmt.Printf("%s : %s(%s)-%s\n", field, vv, vt, tag)
			if val := cb(field, "int", rt.Field(i).Tag, chain); val != "" {
				if a, err := strconv.ParseInt(val, 10, 64); err == nil {
					rv.Field(i).SetInt(a)
					continue
				}
				fmt.Printf("Invalid '%v' while setting value of %s\n", val, field)
				os.Exit(1)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// fmt.Printf("%s : %s(%s)-%s\n", field, vv, vt, tag)
			if val := cb(field, "uint", rt.Field(i).Tag, chain); val != "" {
				base := 10
				if strings.HasPrefix(val, "0x") {
					base = 16
					val = val[2:]
				}
				if strings.HasPrefix(val, "0o") {
					base = 8
					val = val[2:]
				}
				if strings.HasPrefix(val, "0b") {
					base = 2
					val = val[2:]
				}
				if a, err := strconv.ParseUint(val, base, 64); err == nil {
					rv.Field(i).SetUint(a)
					continue
				}
				fmt.Printf("Invalid '%v' while setting value of %s\n", val, field)
				os.Exit(1)
			}
		case reflect.Float32, reflect.Float64:
			// fmt.Printf("%s : %s(%s)-%s\n", field, vv, vt, tag)
			if val := cb(field, "float", rt.Field(i).Tag, chain); val != "" {
				if a, err := strconv.ParseFloat(val, 64); err == nil {
					rv.Field(i).SetFloat(a)
					continue
				}
				fmt.Printf("Invalid '%v' while setting value of %s\n", val, field)
				os.Exit(1)
			}
		case reflect.Bool:
			// fmt.Printf("%s : %s(%s)-%s\n", field, vv, vt, tag)
			if val := cb(field, "bool", rt.Field(i).Tag, chain); val != "" {
				if a, err := strconv.ParseBool(val); err == nil {
					rv.Field(i).SetBool(a)
					continue
				}
				fmt.Printf("Invalid '%v' while setting value of %s\n", val, field)
				os.Exit(1)
			}
		case reflect.Struct:
			// fmt.Printf("%s : it is %s\n", field, vt)
			var step []string
			copy(step, chain)
			step = append(step, field)
			scan(vv.Interface(), step, cb)

			// default:
			// 	fmt.Println("ignoring: ", vt)
		}
	}
}

// LoadDefaults will set all default to the structure fields
func LoadDefaults(c any) {

	cb := func(name, kind string, tags reflect.StructTag, chain []string) string {
		return tags.Get("default")
	}

	scan(c, []string{}, cb)
}

// LoadTOML set structure value to the TOML file content
func LoadTOML(c any, file string) {
	cnt, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read TOML file: %s\n", file)
		os.Exit(1)
	}

	toml.Unmarshal(cnt, c)
}

// LoadEnv set fields to a value define by a environment variable
// we test variables named after prefix + (tag "env" or field name)
func LoadEnv(c any, prefix string, useName bool) {
	cb := func(name, kind string, tags reflect.StructTag, chain []string) string {
		if k := tags.Get("env"); k != "" && k != "-" {
			return os.Getenv(strings.ToUpper(prefix + k))
		}

		if useName {
			n := strings.Join(chain, "_")
			if n != "" {
				n += "_"
			}

			return os.Getenv(strings.ToUpper(prefix + n + name))
		}

		return ""
	}

	scan(c, []string{}, cb)
}

// LoadFlags set structure with values from the command line
func LoadFlags(c any, useName bool) {

	bools, vals, _ := ParseCommand(os.Args[1:])

	cb := func(name, kind string, tags reflect.StructTag, chain []string) string {

		if k := tags.Get("short"); k != "" {

			if bools[k] {
				return "true"
			}

			if v := vals[k]; v != "" {
				return v
			}
		}

		if k := tags.Get("long"); k != "-" {
			if k != "" {

				if bools[k] {
					return "true"
				}

				if v := vals[k]; v != "" {
					return v
				}
			}

			if useName {
				n := strings.Join(chain, "-")
				if n != "" {
					n += "-"
				}
				n += name
				n = strings.ToLower(n)

				if bools[n] {
					return "true"
				}

				if v := vals[n]; v != "" {
					return v
				}
			}
		}

		return ""
	}

	scan(c, []string{}, cb)
}

// GenerateHelp returns a string with help information
func GenerateHelp(c any, prefix string, useName, showShort, showLong, showEnv bool) string {

	type op struct {
		help  string
		short string
		long  string
		env   string
		kind  string
		def   string
		name  string
	}

	ops := []op{}

	cb := func(name, kind string, tags reflect.StructTag, chain []string) string {

		h := op{}

		h.kind = kind
		h.help = tags.Get("help")
		h.def = tags.Get("default")
		h.name = tags.Get("name")

		if showShort {
			h.short = tags.Get("short")
		}
		if showLong {
			h.long = tags.Get("long")
		}
		if h.long == "-" {
			h.long = ""
		} else if useName && h.short == "" && h.long == "" {
			n := strings.Join(chain, "-")
			if n != "" {
				n += "-"
			}
			h.long = strings.ToLower(n + name)
		}

		if e := tags.Get("env"); e != "-" {
			if showEnv && e != "" {
				h.env = strings.ToUpper(prefix + e)
			}

			if useName && h.env == "" {
				n := strings.Join(chain, "_")
				if n != "" {
					n += "_"
				}
				h.env = strings.ToUpper(prefix + n + name)
			}
		}

		if h.short != "" || h.long != "" || h.env != "" {
			ops = append(ops, h)
		}

		return ""
	}

	scan(c, []string{}, cb)

	help := ""

	for _, o := range ops {

		t := []string{}

		if o.short != "" && o.kind == "bool" {
			t = append(t, fmt.Sprintf("-%s", o.short))
		} else if o.short != "" {
			t = append(t, fmt.Sprintf("-%s <%s>", o.short, o.name))
		}

		if o.long != "" && o.kind == "bool" {
			t = append(t, fmt.Sprintf("--%s", o.long))
		} else if o.long != "" {
			t = append(t, fmt.Sprintf("--%s <%s>", o.long, o.name))
		}

		if o.env != "" && o.env != "-" {
			t = append(t, fmt.Sprintf("%s=\"%s\"", o.env, o.name))
		}

		if o.def != "" {
			help += fmt.Sprintf("   %s (%s==\"%s\")\n        %s\n", strings.Join(t, ", "), o.kind, o.def, o.help)
		} else {
			help += fmt.Sprintf("   %s (%s)\n        %s\n", strings.Join(t, ", "), o.kind, o.help)
		}

	}

	return help
}

// IsAskingForHelp return true in case command line includes -h or --help
func IsAskingForHelp() bool {
	b, _, _ := ParseCommand(os.Args[1:])
	return b["h"] || b["help"]
}

// GetArg will return the value of a short line argument -c=VAL, -c by itself is true or empty
func GetArg(name string) string {
	b, m, _ := ParseCommand(os.Args[1:])
	for k := range b {
		if k == name {
			return "true"
		}
	}
	for k, v := range m {
		if k == name {
			return v
		}
	}
	return ""
}

// ParseCommand will parse an []string to brake it into a list of booleans, a list of arguments and
// a map of key:value
func ParseCommand(in []string) (bools map[string]bool, vals map[string]string, args []string) {

	i := 0

	bools = make(map[string]bool)
	vals = make(map[string]string)

	for {
		if i >= len(in) {
			break
		}

		if in[i] == "--" {
			i++
			for i < len(in) {
				args = append(args, in[i])
				i++
			}
			break
		}

		if strings.HasPrefix(in[i], "--") {
			if (i+1) >= len(in) || strings.HasPrefix(in[i+1], "-") {
				bools[in[i][2:]] = true
			} else {
				vals[in[i][2:]] = in[i+1]
				i++
			}
		} else if strings.HasPrefix(in[i], "-") {
			for j, f := range in[i][1:] {
				if j == (len(in[i]) - 2) {
					if (i+1) >= len(in) || strings.HasPrefix(in[i+1], "-") {
						bools[string(f)] = true
					} else {
						vals[string(f)] = in[i+1]
						i++
					}
					break
				}
				bools[string(f)] = true
			}
		} else {
			args = append(args, in[i])
		}

		i++
	}

	// fmt.Println(bools, vals, args)

	return bools, vals, args
}
