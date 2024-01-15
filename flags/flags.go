package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/superwhys/goutils/internal/shared"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service/finder"
	"github.com/superwhys/goutils/service/finder/consul"
	"github.com/superwhys/goutils/slices"

	// Import remote config

	_ "github.com/spf13/viper/remote"
)

type remoteServices struct {
	Services map[string]finder.Service
}

var (
	allKeys           []string
	requiredKey       []string
	nestedKey         = map[string]interface{}{}
	config            *string
	defaultConfigFile string
	debug             *bool

	manualServices *remoteServices

	v = viper.New()

	servicesCache = Struct("remoteService", (*remoteServices)(nil), "Remote service config")
)

func OverrideDefaultConfigFile(configFile string) {
	defaultConfigFile = configFile
}

func initFlags() {
	v.AddConfigPath(".")
	v.AddConfigPath("./tmp/config/")
	err := v.BindPFlags(pflag.CommandLine)
	if err != nil {
		lg.Fatal("BindPFlags Error!")
	}
	shared.PtrServiceName = pflag.String("service", os.Getenv("SERVICE"), "Service name to access the config in remote consul KV store.")
	shared.PtrConsulAddr = pflag.String("consulAddr", consul.HostAddress+":8500", "Consul address")

	config = pflag.StringP("config", "f", defaultConfigFile, "Specify config file to parse. Support json, yaml, toml etc.")
	debug = pflag.Bool("debug", false, "Set true to enable debug mode")

	allKeys = append(allKeys, "debug")
}

func Viper() *viper.Viper {
	return v
}

func GetServiceName() string {
	return shared.GetServiceName()
}

// Parse has to called after main() before any application code.
func Parse() {
	initFlags()
	pflag.Parse()
	finder.SetConsulFinderToDefault()

	if *debug {
		lg.EnableDebug()
	}

	for key, valuePtr := range nestedKey {
		flag := pflag.Lookup(key)
		if flag != nil && flag.Changed {
			switch valuePtr.(type) {
			case *int:
				v.Set(key, *valuePtr.(*int))
			case *bool:
				v.Set(key, *valuePtr.(*bool))
			case *float64:
				v.Set(key, *valuePtr.(*float64))
			case *time.Duration:
				v.Set(key, *valuePtr.(*time.Duration))
			case *string:
				v.Set(key, *valuePtr.(*string))
			case *[]bool:
				v.Set(key, *valuePtr.(*[]bool))
			case *[]string:
				v.Set(key, *valuePtr.(*[]string))
			case *[]int:
				v.Set(key, *valuePtr.(*[]int))
			case *[]float64:
				v.Set(key, *valuePtr.(*[]float64))
			case *[]time.Duration:
				v.Set(key, *valuePtr.(*[]time.Duration))
			default:
				lg.Fatal("Unsupport flag value type", flag.Value.Type())
			}
		}
	}

	if config != nil && *config != "" {
		v.SetConfigFile(*config)
		if err := v.ReadInConfig(); err != nil {
			lg.Error(fmt.Sprintf("Failed to read on local file, ", err))
		} else {
			lg.Info(fmt.Sprintf("Read config from local file: %v!", *config))
		}
	}

	for _, k := range requiredKey {
		if isZero(v.Get(k)) {
			lg.Fatal("Missing", k)
		}
	}
	expectedKeys := slices.NewStringSet(nil)
	for _, k := range allKeys {
		if err := expectedKeys.Add(strings.ToLower(k)); err != nil {
			lg.Fatal(fmt.Sprintf("Add Key Error: --%s", k))
		}
	}

	for _, k := range v.AllKeys() {
		if strings.Contains(k, ".") {
			// Ignore nested key
			continue
		}
		if !expectedKeys.Contains(k) {
			lg.Fatal(fmt.Sprintf("Unknown flag in config: --%s", k))
		}
	}

	if v.GetBool("debug") {
		lg.EnableDebug()
	}

	ms := &remoteServices{}
	lg.PanicError(servicesCache(ms))
	registerManualServiceFinder(ms)
}

func registerManualServiceFinder(ms *remoteServices) {
	for serviceName, serviceInfo := range ms.Services {
		finder.GetServiceFinder().RegisterServiceWithTag(serviceName, serviceInfo.Address, serviceInfo.Tags)
		lg.Debug(fmt.Sprintf("register manual service info, name: %s, address: %s, tags: %v", serviceName, serviceInfo.Address, serviceInfo.Tags))
	}
}

func isZero(i interface{}) bool {
	switch i.(type) {
	case bool:
		// It's trivial to check a bool, since it makes the flag no sense(always true).
		return !i.(bool)
	case string:
		return i.(string) == ""
	case time.Duration:
		return i.(time.Duration) == 0
	case float64:
		return i.(float64) == 0
	case int:
		return i.(int) == 0
	case []string:
		return len(i.([]string)) == 0
	case []interface{}:
		return len(i.([]interface{})) == 0
	default:
		return true
	}
}

func String(key, defaultValue, usage string) func() string {
	pflag.String(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() string {
		return v.GetString(key)
	}
}

func StringRequired(key, usage string) func() string {
	requiredKey = append(requiredKey, key)
	allKeys = append(allKeys, key)
	return String(key, "", usage)
}

func Bool(key string, defaultValue bool, usage string) func() bool {
	pflag.Bool(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() bool {
		return v.GetBool(key)
	}
}

func BoolRequired(key, usage string) func() bool {
	requiredKey = append(requiredKey, key)
	allKeys = append(allKeys, key)
	return Bool(key, false, usage)
}

func Int(key string, defaultValue int, usage string) func() int {
	pflag.Int(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() int {
		return v.GetInt(key)
	}
}

func IntRequired(key, usage string) func() int {
	requiredKey = append(requiredKey, key)
	allKeys = append(allKeys, key)
	return Int(key, 0, usage)
}

func Slice(key string, defaultValue []string, usage string) func() []string {
	pflag.StringSlice(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() []string {
		return v.GetStringSlice(key)
	}
}

func Float64(key string, defaultValue float64, usage string) func() float64 {
	pflag.Float64(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() float64 {
		return v.GetFloat64(key)
	}
}

func Float64Required(key, usage string) func() float64 {
	requiredKey = append(requiredKey, key)
	allKeys = append(allKeys, key)
	return Float64(key, 0, usage)
}

func Duration(key string, defaultValue time.Duration, usage string) func() time.Duration {
	pflag.Duration(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	err := v.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		lg.Fatal(fmt.Sprintf("BindPFlag err, Key: --%s", key))
	}
	allKeys = append(allKeys, key)
	return func() time.Duration {
		return v.GetDuration(key)
	}
}

func DurationRequired(key, usage string) func() time.Duration {
	requiredKey = append(requiredKey, key)
	allKeys = append(allKeys, key)
	return Duration(key, 0, usage)
}

type HasDefault interface {
	SetDefault()
}

type HasValidator interface {
	Validate() error
}

func Struct(key string, defaultValue interface{}, usage string) func(out interface{}) error {
	if err := setPFlagRecursively(key, defaultValue); err != nil {
		lg.Debug("Ignore flag key", key, err)
	}

	v.SetDefault(key, defaultValue)
	allKeys = append(allKeys, key)
	return func(out interface{}) error {
		// structDecodeOpt := viper.DecodeHook(decodeHookFunc())
		if err := v.UnmarshalKey(key, out); err != nil {
			return err
		}
		d, ok := out.(HasDefault)
		if ok {
			d.SetDefault()
		}
		v, ok := out.(HasValidator)
		if ok {
			return v.Validate()
		}
		return nil
	}
}

func setPFlag(key string, ptr interface{}) {
	v.BindPFlag(key, pflag.Lookup(key))
	nestedKey[key] = ptr
}

func setPFlagRecursively(prefix string, i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("Not struct")
	}
	var names []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		name := field.Name
		for _, tag := range []string{"flags", "flag", "json", "bson", "mapstructure"} {
			if content := field.Tag.Get(tag); content != "" {
				name = strings.SplitN(content, ",", 2)[0]
				break
			}
		}
		desc := field.Tag.Get("desc")
		name = prefix + "." + name
		names = append(names, name)

		switch v.Field(i).Kind() {
		case reflect.Bool:
			setPFlag(name, pflag.Bool(name, v.Field(i).Bool(), desc))
		case reflect.Int, reflect.Int64:
			if field.Type.String() == "time.Duration" {
				setPFlag(name, pflag.Duration(name, time.Duration(v.Field(i).Int()), desc))
			} else {
				setPFlag(name, pflag.Int(name, int(v.Field(i).Int()), desc))
			}
		case reflect.Float64:
			setPFlag(name, pflag.Float64(name, v.Field(i).Float(), desc))
		case reflect.String:
			setPFlag(name, pflag.String(name, v.Field(i).String(), desc))
		case reflect.Slice:
			switch field.Type.String() {
			case "[]int":
				setPFlag(name, pflag.IntSlice(name, v.Field(i).Interface().([]int), desc))
			case "[]string":
				setPFlag(name, pflag.StringSlice(name, v.Field(i).Interface().([]string), desc))
			case "[]float64":
				setPFlag(name, pflag.Float64Slice(name, v.Field(i).Interface().([]float64), desc))
			case "[]bool":
				setPFlag(name, pflag.BoolSlice(name, v.Field(i).Interface().([]bool), desc))
			case "[]time.Duration":
				setPFlag(name, pflag.DurationSlice(name, v.Field(i).Interface().([]time.Duration), desc))
			default:
				return fmt.Errorf("Unsupport type of field %s %s", field.Name, field.Type.String())
			}
		case reflect.Struct, reflect.Ptr:
			if err := setPFlagRecursively(name, v.Field(i).Interface()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unsupport kind of field %s %s", field.Name, v.Field(i).Kind())
		}
	}
	return nil
}

func decodeHookFunc() mapstructure.DecodeHookFunc {
	// We use lazy function init to decouple cycle dependency.
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
}
