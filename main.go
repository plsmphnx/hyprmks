package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type (
	Variable struct {
		Name  string
		Value string
	}

	Variables []*Variable

	Flags byte

	Bind struct {
		Flags    string
		Key      string
		Dispatch string
	}

	Binds []*Bind

	Submap struct {
		Alias string
		Binds Binds
	}

	Submaps map[Flags]*Submap

	Modifier struct {
		Name []string
		Flag Flags
		Keys []string
	}
)

var (
	varRx     = regexp.MustCompile(`^\s*(\$\w+)\s*=(.*)`)
	bindRx    = regexp.MustCompile(`^\s*bind([lrenmt]*)\s*=([^,]*),([^,]*),(.*)`)
	aliasRx   = regexp.MustCompile(`^\s*#alias\s*=([^,]*),(.*)`)
	submapRx  = regexp.MustCompile(`^\s*submap\s*=`)
	resetRx   = regexp.MustCompile(`^\s*submap\s*=\s*reset`)
	modifiers = []*Modifier{{
		Name: []string{"SHIFT"},
		Flag: 0b00000001,
		Keys: []string{"shift_l", "shift_r"},
	}, {
		Name: []string{"CAPS"},
		Flag: 0b00000010,
		Keys: []string{"caps_lock"},
	}, {
		Name: []string{"CTRL", "CONTROL"},
		Flag: 0b00000100,
		Keys: []string{"control_l", "control_r"},
	}, {
		Name: []string{"ALT"},
		Flag: 0b00001000,
		Keys: []string{"alt_l", "alt_r"},
	}, {
		Name: []string{"MOD2"},
		Flag: 0b00010000,
		Keys: []string{},
	}, {
		Name: []string{"MOD3"},
		Flag: 0b00100000,
		Keys: []string{},
	}, {
		Name: []string{"SUPER", "WIN", "LOGO", "MOD4"},
		Flag: 0b01000000,
		Keys: []string{"super_l", "super_r"},
	}, {
		Name: []string{"MOD5"},
		Flag: 0b10000000,
		Keys: []string{},
	}}
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <hyprland.conf>\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var vars Variables
	submaps := make(Submaps)
	var ignore bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if ignore {
			if resetRx.MatchString(line) {
				ignore = false
			}
			continue
		}

		if submapRx.MatchString(line) {
			ignore = true
			continue
		}

		if m := varRx.FindStringSubmatch(line); m != nil {
			vars = append(vars, &Variable{m[1], strings.TrimSpace(m[2])})
			continue
		}

		if m := bindRx.FindStringSubmatch(line); m != nil {
			submap := submaps.Get(vars.Apply(m[2]))
			submap.Binds = append(submap.Binds, &Bind{
				strings.TrimSpace(m[1]),
				strings.TrimSpace(m[3]),
				strings.TrimSpace(m[4]),
			})
			continue
		}

		if m := aliasRx.FindStringSubmatch(line); m != nil {
			submaps.Get(vars.Apply(m[1])).Alias = strings.TrimSpace(m[2])
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	submaps.Print()
}

func (v Variables) Apply(str string) string {
	res := str
	for i := len(v) - 1; i >= 0; i-- {
		res = strings.ReplaceAll(res, v[i].Name, v[i].Value)
	}
	return res
}

func (s Submaps) Get(mods string) *Submap {
	var flags Flags
	for _, mod := range modifiers {
		for _, name := range mod.Name {
			if strings.Contains(mods, name) {
				flags |= mod.Flag
			}
		}
	}

	if submap, ok := s[flags]; ok {
		return submap
	}
	submap := &Submap{Alias: strings.TrimSpace(mods)}
	s[flags] = submap
	return submap
}

func (s Submaps) Print() {
	order := make([]Flags, 0, len(s))
	for flags := range s {
		order = append(order, flags)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })

	if order[0] == 0 {
		order = order[1:]
	}

	for i, flags := range order {
		submap := s[flags]
		flags.PrintEnter(submap.Alias)

		fmt.Printf("\nsubmap=%s\n", submap.Alias)

		flags.PrintExit()

		submap.Binds.Print(0, true)

		for _, next := range order[i+1:] {
			if flags&next == flags {
				child := s[next]
				diff := next &^ flags

				diff.PrintEnter(child.Alias)
				child.Binds.Print(diff, false)
			}
		}

		fmt.Printf("\nsubmap=reset\n")
	}
}

func (f Flags) String() string {
	var mods []string
	for _, mod := range modifiers {
		if f&mod.Flag != 0 {
			mods = append(mods, mod.Name[0])
		}
	}
	return strings.Join(mods, "_")
}

func (f Flags) Mods() []*Modifier {
	var mods []*Modifier
	for _, mod := range modifiers {
		if f&mod.Flag != 0 {
			mods = append(mods, mod)
		}
	}
	return mods
}

func (f Flags) PrintEnter(submap string) {
	fmt.Printf("\n")
	mods := f.String()
	for _, mod := range f.Mods() {
		for _, key := range mod.Keys {
			fmt.Printf("bindr=%s,%s,submap,%s\n", mods, key, submap)
		}
	}
}

func (f Flags) PrintExit() {
	fmt.Printf("\n")
	for _, mod := range f.Mods() {
		for _, key := range mod.Keys {
			fmt.Printf("bindr=%s,%s,submap,reset\n", mod.Name[0], key)
		}
	}
}

func (b Binds) Print(flags Flags, reset bool) {
	if len(b) > 0 {
		fmt.Printf("\n")
	}

	mods := flags.String()
	for _, bind := range b {
		fmt.Printf(
			"bind%s=%s,%s,%s\n",
			bind.Flags,
			mods,
			bind.Key,
			bind.Dispatch,
		)
		if reset {
			fmt.Printf(
				"bind%s=%s,%s,submap,reset\n",
				bind.Flags,
				mods,
				bind.Key,
			)
		}
	}
}
