package oneconf

import (
	"fmt"
	"os"
)

func ExampleLoadDefaults() {

	type cfg struct {
		A bool
		B bool    `default:"t"`
		C bool    `default:"1"`
		D bool    `default:"true"`
		E bool    `default:"false"`
		F int8    `default:"12"`
		G int32   `default:"-12"`
		H uint8   `default:"70"`
		I uint16  `default:"70"`
		J uint32  `default:"0x70"`
		K uint64  `default:"0o70"`
		L uint    `default:"0b10011"`
		M float32 `default:"12e-3"`
		N float64 `default:"12e+1"`
		O string  `default:"text"`
	}

	c := cfg{}

	LoadDefaults(&c)

	fmt.Println(c)

	// Output: {false true true true false 12 -12 70 70 112 56 19 0.012 120 text}
}

func ExampleLoadEnv() {
	type cfg struct {
		A bool
		B bool `env:"one"`
		C bool `env:"two"`
		D bool
		E bool
		F int8
		G int32
		H uint8
		I uint16
		J uint32 `env:"THREE"`
		K uint64
		L uint
		M float32
		N float64
		O string
	}

	os.Setenv("PRE_ONE", "true")
	os.Setenv("PRE_TWO", "T")
	os.Setenv("PRE_THREE", "0x77") // used to populate J
	os.Setenv("PRE_D", "T")
	os.Setenv("PRE_E", "F")
	os.Setenv("PRE_F", "6")
	os.Setenv("PRE_G", "-8")
	os.Setenv("PRE_H", "110")
	os.Setenv("PRE_I", "0x110")
	os.Setenv("PRE_J", "0o110") //ignored since field has a env name
	os.Setenv("PRE_K", "0b110")
	os.Setenv("PRE_L", "110")
	os.Setenv("PRE_M", "1.28e-2")
	os.Setenv("PRE_N", "3.1415")
	os.Setenv("PRE_O", "my-text")

	c := cfg{}

	LoadEnv(&c, "PRE_", false) // populate named ones only
	fmt.Println(c)

	LoadEnv(&c, "PRE_", true) // populate named ones and then by field name
	fmt.Println(c)

	// Output:
	// {false true true false false 0 0 0 0 119 0 0 0 0 }
	// {false true true true false 6 -8 110 272 119 6 110 0.0128 3.1415 my-text}
}

func ExampleLoadTOML() {

	type cfg struct {
		A string
		B int
		C []string
		D struct {
			E bool
			F float64
		}
	}

	toml := `
A = "my-string"
B = 0x110
C = ["one", "two"]

[D]
E = true
F = 12e-1
	`

	if err := os.WriteFile("_$_test.toml", []byte(toml), 0o600); err != nil {
		fmt.Println("can not save test file")
		os.Exit(1)
	}

	c := cfg{}

	LoadTOML(&c, "_$_test.toml")

	fmt.Print(c)

	if err := os.Remove("_$_test.toml"); err != nil {
		fmt.Println("can not remove test file")
		os.Exit(1)
	}

	// Output: {my-string 272 [one two] {true 1.2}}
}

func ExampleLoadFlags() {

	type cfg struct {
		A bool
		B bool
		C bool
		D bool
		E bool
		F int8
		G int32
		H uint8
		I uint16
		J uint32
		K uint64
		L uint
		M float32
		N float64
		O string
	}

	c := cfg{}

	os.Args = []string{"binary", "-h"}

	LoadFlags(&c, false) // populate named ones only
	fmt.Println(c)

	LoadFlags(&c, true) // populate named ones and then by field name
	fmt.Println(c)

	// Outputs:
}
