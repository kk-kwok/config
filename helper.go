package config

import (
    "errors"
    "fmt"
    "os"

    "github.com/kk-kwok/config/version"

    v10validator "github.com/go-playground/validator/v10"
)

// DumpDemoCfg dump the config to stdout and exit the app
// nolint: forbidigo
func DumpDemoCfg(cfg interface{}) {
    // print the version
    fmt.Printf("# %s %s\n", version.ServiceName, version.Info())
    text, err := TomlMarshalIndent(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "toml.Marshal failed with error: %v", err)
        os.Exit(2)
    }
    fmt.Fprintln(os.Stdout, text)
    fmt.Fprintln(os.Stderr, "config dump success")
}

func ValidateConfig(cfg interface{}) error {
    validate := v10validator.New()
    err := validate.Struct(cfg)
    if err != nil {
        var validationErrors v10validator.ValidationErrors
        if errors.As(err, &validationErrors) {
            return validationErrors
        }
        return err
    }
    return nil
}
