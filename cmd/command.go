package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"
)

// CheckErr prints the msg with the prefix 'Error:' and exits with error code 1. If the msg is nil, it does nothing.
func CheckErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\033[1;31m%s\033[0m", fmt.Sprintf("Error: %s\n", err.Error()))
		os.Exit(1)
	}
}

func SilenceCmdErrors(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	for _, subCmd := range cmd.Commands() {
		SilenceCmdErrors(subCmd)
	}
}

func AddFlags(cmd *cobra.Command, name string, value interface{}, usage string, required bool) {
	switch v := value.(type) {
	case string:
		cmd.Flags().String(name, v, usage)
	case int64:
		cmd.Flags().Int64(name, v, usage)
	case uint64:
		cmd.Flags().Uint64(name, v, usage)
	case int:
		cmd.Flags().Int(name, v, usage)
	case uint:
		cmd.Flags().Uint(name, v, usage)
	case float64:
		cmd.Flags().Float64(name, v, usage)
	case float32:
		cmd.Flags().Float32(name, v, usage)
	case bool:
		cmd.Flags().Bool(name, v, usage)
	default:
		panic("无效的flag类型")
	}
	if required {
		CheckErr(cmd.MarkFlagRequired(name))
	}
}

func GenFlags(v interface{}, parents ...string) *flag.FlagSet {
	valueOf, ok := v.(reflect.Value)
	if !ok {
		valueOf = reflect.ValueOf(v)
	}
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		return nil
	}
	flagSet := flag.NewFlagSet(valueOf.Type().Name(), flag.ContinueOnError)

	for j := 0; j < valueOf.NumField(); j++ {
		flagUsage := valueOf.Type().Field(j).Tag.Get("flag")
		if flagUsage == "-" {
			continue
		}
		name := valueOf.Type().Field(j).Name
		for i := len(parents) - 1; i >= 0; i-- {
			name = fmt.Sprintf("%s.%s", parents[i], name)
		}
		switch valueOf.Field(j).Type().Kind() {
		case reflect.String:
			flagSet.String(name, valueOf.Field(j).String(), flagUsage)
		case reflect.Int64:
			flagSet.Int64(name, valueOf.Field(j).Int(), flagUsage)
		case reflect.Uint64:
			flagSet.Uint64(name, uint64(valueOf.Field(j).Int()), flagUsage)
		case reflect.Int:
			flagSet.Int(name, int(valueOf.Field(j).Int()), flagUsage)
		case reflect.Uint:
			flagSet.Uint(name, uint(valueOf.Field(j).Int()), flagUsage)
		case reflect.Bool:
			flagSet.Bool(name, valueOf.Field(j).Bool(), flagUsage)
		default:
			GenFlags(valueOf.Field(j), name).VisitAll(func(f *flag.Flag) {
				flagSet.Var(f.Value, f.Name, f.Usage)
			})
		}
	}
	return flagSet
}
