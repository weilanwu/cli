package i18n

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/XenoPhex/jibber_jabber"

	resources "github.com/cloudfoundry/cli/cf/resources"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"
)

const (
	DEFAULT_LOCAL = "en_US"
)

var SUPPORTED_LANGUAGES = []string{"ar", "ca", "zh", "cs", "da", "nl", "en", "fr", "de", "it", "ja", "lt", "pt", "es"}
var resources_path = filepath.Join("cf", "i18n", "resources")

func GetResourcesPath() string {
	return resources_path
}

func Init(packageName string, i18nDirname string) go_i18n.TranslateFunc {
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil {
		println("Could not load desired locale:", userLocale, "falling back to default locale", DEFAULT_LOCAL)
		userLocale = DEFAULT_LOCAL
	}

	// convert IETF format to XCU format
	userLocale = strings.Replace(userLocale, "-", "_", 1)

	err = loadFromAsset(packageName, i18nDirname, userLocale)
	if err != nil { // this should only be from the user locale
		println("Could not load desired locale:", userLocale, "falling back to default locale", DEFAULT_LOCAL)
	}

	T, err := go_i18n.Tfunc(userLocale, DEFAULT_LOCAL)
	if err != nil {
		panic(err)
	}

	return T
}

func splitLocale(locale string) (string, string) {
	formattedLocale := strings.Split(locale, ".")[0]
	formattedLocale = strings.Replace(formattedLocale, "-", "_", -1)
	language := strings.Split(formattedLocale, "_")[0]
	territory := strings.Split(formattedLocale, "_")[1]
	return language, territory
}

func loadFromAsset(packageName, assetPath, locale string) error {
	language, _ := splitLocale(locale)
	assetName := locale + ".all.json"
	assetKey := filepath.Join(assetPath, language, packageName, assetName)

	byteArray, err := resources.Asset(assetKey)
	if err != nil {
		return err
	}

	if len(byteArray) == 0 {
		return errors.New(fmt.Sprintf("Could not load i18n asset: %v", assetKey))
	}

	tmpDir, err := ioutil.TempDir("", "cloudfoundry_cli_i18n_res")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	fileName, err := saveLanguageFileToDisk(tmpDir, assetName, byteArray)
	if err != nil {
		return err
	}

	go_i18n.MustLoadTranslationFile(fileName)

	os.RemoveAll(fileName)

	return nil
}

func saveLanguageFileToDisk(tmpDir, assetName string, byteArray []byte) (fileName string, err error) {
	fileName = filepath.Join(tmpDir, assetName)
	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(byteArray)
	if err != nil {
		return
	}

	return
}
