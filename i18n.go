package main

import (
	"embed"
	"github.com/Xuanwo/go-locale"
	"github.com/fullpipe/icu-mf/mf"
	"golang.org/x/text/language"
)

//go:embed i18n/*.yaml
var i18nDir embed.FS
var TR mf.Translator
var LanguageTag string
var i18nBundle mf.Bundle
var isRefreshingUI bool // 防止UI刷新时的递归调用

func detectSystemLanguate() string {
	tag, err := locale.Detect()
	if err != nil {
		return "en"
	}
	return tag.String()
}

func InitI18n() {
	bundle, err := mf.NewBundle(
		mf.WithDefaultLangFallback(language.English),
		mf.WithYamlProvider(i18nDir))
	if err != nil {
		panic(err)
	}
	i18nBundle = bundle
	
	// 如果Config中有语言设置，使用配置的语言，否则使用系统语言
	if ConfigInstance.Language != "" {
		LanguageTag = ConfigInstance.Language
	} else {
		LanguageTag = detectSystemLanguate()
	}
	
	TR = bundle.Translator(LanguageTag)
}

// SetLanguage 切换语言
// langTag: 语言标签，如 "en", "zh-TW", "ja-JP"
func SetLanguage(langTag string) {
	if i18nBundle == nil {
		return
	}
	LanguageTag = langTag
	ConfigInstance.Language = langTag
	TR = i18nBundle.Translator(LanguageTag)
	RefreshAllUI()
}

// RefreshAllUI 刷新所有UI文本
func RefreshAllUI() {
	if WMain == nil || isRefreshingUI {
		return
	}
	
	isRefreshingUI = true
	defer func() {
		isRefreshingUI = false
	}()
	
	// 刷新所有标签和按钮文本
	StatusLabel.SetText(TR.Trans("label.status_ready"))
	DownloadButton.SetText(TR.Trans("label.download_profile_button"))
	SetNicknameButton.SetText(TR.Trans("label.set_nickname_button"))
	DeleteProfileButton.SetText(TR.Trans("label.delete_profile_button"))
	SwitchStateButton.SetText(TR.Trans("label.switch_state_button_enable"))
	ProcessNotificationButton.SetText(TR.Trans("label.process_notification_button"))
	ProcessAllNotificationButton.SetText(TR.Trans("label.process_all_notification_button"))
	RemoveNotificationButton.SetText(TR.Trans("label.remove_notification_button"))
	BatchRemoveNotificationButton.SetText(TR.Trans("label.batch_remove_notification_button"))
	OpenLogButton.SetText(TR.Trans("label.open_log_button"))
	RefreshButton.SetText(TR.Trans("label.refresh_button"))
	ProfileMaskCheck.SetText(TR.Trans("label.profile_mask_check"))
	NotificationMaskCheck.SetText(TR.Trans("label.notification_mask_check"))
	CopyEidButton.SetText(TR.Trans("label.copy_eid_button"))
	ViewCertInfoButton.SetText(TR.Trans("label.view_cert_info_button"))
	CopyEuiccInfo2Button.SetText(TR.Trans("label.copy_euicc_info2_button"))
	
	// 刷新标签页标题
	ProfileTab.Text = TR.Trans("tab_bar.profile")
	NotificationTab.Text = TR.Trans("tab_bar.notification")
	ChipInfoTab.Text = TR.Trans("tab_bar.chip_info")
	SettingsTab.Text = TR.Trans("tab_bar.settings")
	AboutTab.Text = TR.Trans("tab_bar.about")
	
	// 刷新语言选择框（需要防止递归调用）
	if LanguageSelect != nil {
		languageCodes := []string{"auto", "en", "zh-TW", "ja-JP"}
		languageOptions := []string{
			TR.Trans("label.language_auto"),
			TR.Trans("label.language_en"),
			TR.Trans("label.language_zh_tw"),
			TR.Trans("label.language_ja_jp"),
		}
		currentCode := ConfigInstance.Language
		if currentCode == "" {
			currentCode = "auto"
		}
		
		// 找到当前语言代码对应的索引
		var currentIndex int = 0
		for i, code := range languageCodes {
			if code == currentCode {
				currentIndex = i
				break
			}
		}
		
		// 先更新选项
		LanguageSelect.SetOptions(languageOptions)
		
		// 检查当前选择是否已经是正确的，避免触发回调
		currentSelected := LanguageSelect.Selected
		expectedSelected := languageOptions[currentIndex]
		
		// 只有当选择不同时才更新，避免触发回调导致死循环
		if currentSelected != expectedSelected && currentIndex >= 0 && currentIndex < len(languageOptions) {
			LanguageSelect.SetSelected(languageOptions[currentIndex])
		}
	}
	
	// 刷新列表
	ProfileList.Refresh()
	NotificationList.Refresh()
	
	// 刷新窗口标题
	WMain.SetTitle("EasyLPAC")
	
	// 刷新主题（因为字体可能因语言而改变）
	App.Settings().SetTheme(&MyTheme{})
}
