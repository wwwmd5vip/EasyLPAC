package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

const Version = "development"
const EUICCDataVersion = "unknown"

var App fyne.App

func init() {
	InitCiRegistry()
	InitEumRegistry()
	
	// 先加载配置（因为InitI18n需要读取ConfigInstance.Language）
	if err := LoadConfig(); err != nil {
		panic(err)
	}
	
	// 然后初始化i18n（会读取ConfigInstance.Language）
	InitI18n()
	
	App = app.New()
	App.Settings().SetTheme(&MyTheme{})

	if _, err := os.Stat(ConfigInstance.LogDir); os.IsNotExist(err) {
		err := os.Mkdir(ConfigInstance.LogDir, 0755)
		if err != nil {
			panic(err)
		}
	}
	
	// 加载AID列表
	if aidList, err := LoadAidList(); err == nil {
		AidList = aidList
	}
	// 如果加载失败，不panic，允许程序继续运行（AID列表是可选的）
}

func main() {
	var err error
	ConfigInstance.LogFile, err = os.Create(filepath.Join(ConfigInstance.LogDir, ConfigInstance.LogFilename))
	if err != nil {
		panic(err)
	}
	defer ConfigInstance.LogFile.Close()

	InitWidgets()
	go UpdateStatusBarListener()
	go LockButtonListener()

	WMain = InitMainWindow()

	_, err = os.Stat(filepath.Join(ConfigInstance.LpacDir, ConfigInstance.EXEName))
	if err != nil {
		d := dialog.NewError(fmt.Errorf(" %s",TR.Trans("message.lpac_not_found")), WMain)
		d.SetOnClosed(func() {
			os.Exit(127)
		})
		d.Show()
	} else {
		if version, err2 := LpacVersion(); err2 != nil {
			LpacVersionLabel.SetText(TR.Trans("label.lpac_version_unknown"))
		} else {
			LpacVersionLabel.SetText(TR.Trans("label.lpac_version") + " " + version)
		}
		RefreshApduDriver()
		if ApduDrivers != nil {
			ApduDriverSelect.SetSelectedIndex(0)
		}
	}

	WMain.Show()
	App.Run()
}
