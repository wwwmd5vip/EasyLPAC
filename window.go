package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/makiuchi-d/gozxing"
	nativeDialog "github.com/sqweek/dialog"
	"golang.design/x/clipboard"
)

var WMain fyne.Window
var spacer *canvas.Rectangle

func InitMainWindow() fyne.Window {
	w := App.NewWindow("EasyLPAC")
	w.Resize(fyne.Size{
		Width:  850,
		Height: 545,
	})
	w.SetMaster()

	statusBar := container.NewGridWrap(fyne.Size{
		Width:  100,
		Height: DownloadButton.MinSize().Height,
	}, StatusLabel, StatusProcessBar)

	spacer = canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(1, 1))

	topToolBar := container.NewBorder(
		layout.NewSpacer(),
		nil,
		container.New(layout.NewHBoxLayout(), OpenLogButton, spacer, RefreshButton, spacer),
		FreeSpaceLabel,
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel(TR.Trans("label.card_reader")),
			nil,
			container.NewHBox(container.NewGridWrap(fyne.Size{
				Width:  280,
				Height: ApduDriverSelect.MinSize().Height,
			}, ApduDriverSelect), ApduDriverRefreshButton)),
	)

	profileTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(ProfileMaskCheck, DownloadButton,
				// spacer, DiscoveryButton,
				spacer, SetNicknameButton,
				spacer, SwitchStateButton,
				spacer, DeleteProfileButton),
			statusBar),
		nil,
		nil,
		ProfileList)
	ProfileTab = container.NewTabItem(TR.Trans("tab_bar.profile"), profileTabContent)

	notificationTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(NotificationMaskCheck,
				spacer, ProcessNotificationButton,
				spacer, ProcessAllNotificationButton,
				spacer, BatchRemoveNotificationButton,
				spacer, RemoveNotificationButton),
			statusBar),
		nil,
		nil,
		NotificationList)
	NotificationTab = container.NewTabItem(TR.Trans("tab_bar.notification"), notificationTabContent)

	chipInfoTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			nil,
			statusBar),
		nil,
		nil,
		container.NewBorder(
			container.NewVBox(
				container.NewHBox(
					EidLabel, CopyEidButton, layout.NewSpacer(), EUICCManufacturerLabel),
				container.NewHBox(
					DefaultDpAddressLabel, SetDefaultSmdpButton, layout.NewSpacer(), ViewCertInfoButton),
				container.NewHBox(
					RootDsAddressLabel, layout.NewSpacer(), CopyEuiccInfo2Button)),
			nil,
			nil,
			nil,
			container.NewScroll(EuiccInfo2Entry),
		))
	ChipInfoTab = container.NewTabItem(TR.Trans("tab_bar.chip_info"), chipInfoTabContent)

	aidEntryHint := &widget.Label{Text: TR.Trans("label.aid_valid")}
	aidEntry := &widget.Entry{
		Text: ConfigInstance.LpacAID,
		Validator: validation.NewAllStrings(
			validation.NewRegexp(`^.{32}$`, TR.Trans("message.aid_length_illegal")),
			validation.NewRegexp(`[[:xdigit:]]{32}`, TR.Trans("message.aid_not_hex")),
		),
	}
	aidEntry.OnChanged = func(s string) {
		val := aidEntry.Validate()
		if val != nil {
			aidEntryHint.SetText(val.Error())
		} else {
			// Use last known good value only
			ConfigInstance.LpacAID = s
			aidEntryHint.SetText(TR.Trans("label.aid_valid"))
		}
	}
	setToDefaultAidButton := widget.NewButton(
		TR.Trans("label.aid_default_button"),
		func() {
			aidEntry.SetText(AID_DEFAULT)
		})
	setTo5berAidButton := widget.NewButton(
		TR.Trans("label.aid_5ber_button"),
		func() {
			aidEntry.SetText(AID_5BER)
		})
	setToEsimmeAidButton := widget.NewButton(
		TR.Trans("label.aid_esimme_button"),
		func() {
			aidEntry.SetText(AID_ESIMME)
		})
	setToXesimAidButton := widget.NewButton(
		TR.Trans("label.aid_xesim_button"),
		func() {
			aidEntry.SetText(AID_XESIM)
		})
	
	// AID列表选择按钮
	selectFromAidListButton := widget.NewButton(
		TR.Trans("label.aid_select_from_list_button"),
		func() {
			InitAidListDialog(aidEntry).Show()
		})

	settingsTabContent := container.NewVBox(
		&widget.Label{Text: TR.Trans("label.lpac_isdr_aid"), TextStyle: fyne.TextStyle{Bold: true}},
		container.NewHBox(container.NewGridWrap(
			fyne.Size{
				Width:  320,
				Height: aidEntry.MinSize().Height,
			}, aidEntry),
			setToDefaultAidButton,
			setTo5berAidButton,
			setToEsimmeAidButton,
			setToXesimAidButton,
			selectFromAidListButton),
		aidEntryHint,

		&widget.Label{Text: TR.Trans("label.lpac_debug_output"), TextStyle: fyne.TextStyle{Bold: true}},
		&widget.Check{
			Text:    TR.Trans("label.enable_env_LIBEUICC_DEBUG_HTTP_check"),
			Checked: false,
			OnChanged: func(b bool) {
				ConfigInstance.DebugHTTP = b
			},
		},
		&widget.Check{
			Text:    TR.Trans("label.enable_env_LIBEUICC_DEBUG_APDU_check"),
			Checked: false,
			OnChanged: func(b bool) {
				ConfigInstance.DebugAPDU = b
			},
		},

		&widget.Label{Text: TR.Trans("label.easylpac_settings"), TextStyle: fyne.TextStyle{Bold: true}},
		&widget.Check{
			Text:    TR.Trans("label.auto_process_notification_check"),
			Checked: true,
			OnChanged: func(b bool) {
				ConfigInstance.AutoMode = b
			},
		},
		
		&widget.Label{Text: TR.Trans("label.language_settings"), TextStyle: fyne.TextStyle{Bold: true}},
		container.NewHBox(
			widget.NewLabel(TR.Trans("label.language")),
			func() *widget.Select {
				// 语言代码列表（按顺序）
				languageCodes := []string{"auto", "en", "zh-TW", "ja-JP"}
				
				// 获取当前语言的选项文本
				getLanguageOptions := func() []string {
					return []string{
						TR.Trans("label.language_auto"),
						TR.Trans("label.language_en"),
						TR.Trans("label.language_zh_tw"),
						TR.Trans("label.language_ja_jp"),
					}
				}
				
				options := getLanguageOptions()
				languageSelect := widget.NewSelect(options, func(selectedText string) {
					// 根据选择的文本找到对应的索引
					options := getLanguageOptions()
					var selectedIndex int = -1
					for i, opt := range options {
						if opt == selectedText {
							selectedIndex = i
							break
						}
					}
					
					if selectedIndex >= 0 && selectedIndex < len(languageCodes) {
						langCode := languageCodes[selectedIndex]
						
						// 根据语言代码切换语言
						if langCode == "auto" {
							ConfigInstance.Language = ""
							LanguageTag = detectSystemLanguate()
							TR = i18nBundle.Translator(LanguageTag)
						} else {
							SetLanguage(langCode)
						}
					}
				})
				
				// 设置当前选择
				currentCode := ConfigInstance.Language
				if currentCode == "" {
					currentCode = "auto"
				}
				for i, code := range languageCodes {
					if code == currentCode {
						options := getLanguageOptions()
						if i < len(options) {
							languageSelect.SetSelected(options[i])
						}
						break
					}
				}
				
				// 保存到全局变量
				LanguageSelect = languageSelect
				
				return languageSelect
			}(),
		))
	SettingsTab = container.NewTabItem(TR.Trans("tab_bar.settings"), settingsTabContent)

	thankstoText := widget.NewRichTextFromMarkdown(TR.Trans("thanks_to"))

	aboutText := widget.NewRichTextFromMarkdown(TR.Trans("about"))

	aboutTabContent := container.NewBorder(
		nil,
		container.NewBorder(nil, nil,
			container.NewHBox(
				widget.NewLabel(fmt.Sprintf(TR.Trans("label.version")+" %s", Version)),
				LpacVersionLabel),
			widget.NewLabel(fmt.Sprintf(TR.Trans("label.euicc_data")+" %s", EUICCDataVersion))),
		nil,
		nil,
		container.NewCenter(container.NewVBox(thankstoText, aboutText)))
	AboutTab = container.NewTabItem(TR.Trans("tab_bar.about"), aboutTabContent)

	Tabs = container.NewAppTabs(ProfileTab, NotificationTab, ChipInfoTab, SettingsTab, AboutTab)

	w.SetContent(Tabs)

	return w
}

func InitDownloadDialog() dialog.Dialog {
	smdpEntry := &widget.Entry{PlaceHolder: TR.Trans("label.smdp_entry_placeholder")}
	matchIDEntry := &widget.Entry{PlaceHolder: TR.Trans("label.match_id_entry_placeholder")}
	confirmCodeEntry := &widget.Entry{PlaceHolder: TR.Trans("label.confirm_code_entry_placeholder")}
	imeiEntry := &widget.Entry{PlaceHolder: TR.Trans("label.imei_entry_placeholder")}

	formItems := []*widget.FormItem{
		{Text: TR.Trans("label.smdp"), Widget: smdpEntry},
		{Text: TR.Trans("label.match_id"), Widget: matchIDEntry},
		{Text: TR.Trans("label.confirm_code"), Widget: confirmCodeEntry},
		{Text: TR.Trans("label.imei"), Widget: imeiEntry},
	}

	form := widget.NewForm(formItems...)
	var d dialog.Dialog
	showConfirmCodeNeededDialog := func() {
		dialog.ShowInformation(TR.Trans("dialog.confirm_code_required"),
			TR.Trans("message.confirm_code_required"), WMain)
	}
	cancelButton := &widget.Button{
		Text: TR.Trans("dialog.cancel"),
		Icon: theme.CancelIcon(),
		OnTapped: func() {
			d.Hide()
		},
	}
	downloadButton := &widget.Button{
		Text:       TR.Trans("label.download_profile_button"),
		Icon:       theme.ConfirmIcon(),
		Importance: widget.HighImportance,
		OnTapped: func() {
			d.Hide()
			pullConfig := PullInfo{
				SMDP:        strings.TrimSpace(smdpEntry.Text),
				MatchID:     strings.TrimSpace(matchIDEntry.Text),
				ConfirmCode: strings.TrimSpace(confirmCodeEntry.Text),
				IMEI:        strings.TrimSpace(imeiEntry.Text),
			}
			go func() {
				err := RefreshNotification()
				if err != nil {
					ShowLpacErrDialog(err)
					return
				}
				LpacProfileDownload(pullConfig)
			}()
		},
	}
	// 回调函数需要操作这两个 Button，预先声明
	var selectQRCodeButton *widget.Button
	var pasteFromClipboardButton *widget.Button
	disableButtons := func() {
		cancelButton.Disable()
		downloadButton.Disable()
		selectQRCodeButton.Disable()
		pasteFromClipboardButton.Disable()
	}
	enableButtons := func() {
		cancelButton.Enable()
		downloadButton.Enable()
		selectQRCodeButton.Enable()
		pasteFromClipboardButton.Enable()
	}

	selectQRCodeButton = &widget.Button{
		Text: TR.Trans("label.select_qrcode_button"),
		Icon: theme.FileImageIcon(),
		OnTapped: func() {
			go func() {
				disableButtons()
				defer enableButtons()
				fileBuilder := nativeDialog.File().Title(TR.Trans("dialog.select_qrcode"))
				fileBuilder.Filters = []nativeDialog.FileFilter{
					{
						Desc:       TR.Trans("dialog.image_desc") + " (*.PNG, *.png, *.JPG, *.jpg, *.JPEG, *.jpeg)",
						Extensions: []string{"PNG", "png", "JPG", "jpg", "JPEG", "jpeg"},
					},
					{
						Desc:       TR.Trans("dialog.all_files_desc") + " (*.*)",
						Extensions: []string{"*"},
					},
				}

				filename, err := fileBuilder.Load()
				if err != nil {
					if err.Error() != "Cancelled" {
						panic(err)
					}
				} else {
					result, err := ScanQRCodeImageFile(filename)
					if err != nil {
						dialog.ShowError(err, WMain)
					} else {
						pullInfo, confirmCodeNeeded, err2 := DecodeLpaActivationCode(result.String())
						if err2 != nil {
							dialog.ShowError(err2, WMain)
						} else {
							smdpEntry.SetText(pullInfo.SMDP)
							matchIDEntry.SetText(pullInfo.MatchID)
							if confirmCodeNeeded {
								go showConfirmCodeNeededDialog()
							}
						}
					}
				}
			}()
		},
	}
	pasteFromClipboardButton = &widget.Button{
		Text: TR.Trans("label.paste_from_clipboard_button"),
		Icon: theme.ContentPasteIcon(),
		OnTapped: func() {
			go func() {
				disableButtons()
				defer enableButtons()
				var err error
				var pullInfo PullInfo
				var confirmCodeNeeded bool
				var qrResult *gozxing.Result

				format, result, err := PasteFromClipboard()
				if err != nil {
					dialog.ShowError(err, WMain)
					return
				}
				switch format {
				case clipboard.FmtImage:
					qrResult, err = ScanQRCodeImageBytes(result)
					if err != nil {
						dialog.ShowError(err, WMain)
						return
					}
					pullInfo, confirmCodeNeeded, err = DecodeLpaActivationCode(qrResult.String())
				case clipboard.FmtText:
					pullInfo, confirmCodeNeeded, err = DecodeLpaActivationCode(CompleteActivationCode(string(result)))
				default:
					// Unreachable, should not be here.
					panic("unexpected clipboard format")
				}
				if err != nil {
					dialog.ShowError(err, WMain)
					return
				}
				smdpEntry.SetText(pullInfo.SMDP)
				matchIDEntry.SetText(pullInfo.MatchID)
				if confirmCodeNeeded {
					go showConfirmCodeNeededDialog()
				}
			}()
		},
	}
	d = dialog.NewCustomWithoutButtons(TR.Trans("label.download_profile_button"), container.NewBorder(
		nil,
		container.NewVBox(spacer, container.NewCenter(selectQRCodeButton), spacer,
			container.NewCenter(pasteFromClipboardButton), spacer,
			container.NewCenter(container.NewHBox(cancelButton, spacer, downloadButton))),
		nil,
		nil,
		form), WMain)
	d.Resize(fyne.Size{
		Width:  520,
		Height: 380,
	})
	return d
}

func InitSetNicknameDialog() dialog.Dialog {
	entry := &widget.Entry{PlaceHolder: TR.Trans("label.set_nickname_entry_placeholder")}
	form := []*widget.FormItem{
		{Text: TR.Trans("label.set_nickname_button"), Widget: entry},
	}
	d := dialog.NewForm(TR.Trans("label.set_nickname_form"), TR.Trans("dialog.submit"), TR.Trans("dialog.cancel"), form, func(b bool) {
		if b {
			if err := LpacProfileNickname(Profiles[SelectedProfile].Iccid, entry.Text); err != nil {
				ShowLpacErrDialog(err)
			}
			err := RefreshProfile()
			if err != nil {
				ShowLpacErrDialog(err)
			}
		}
	}, WMain)
	d.Resize(fyne.Size{
		Width:  400,
		Height: 180,
	})
	return d
}

func InitSetDefaultSmdpDialog() dialog.Dialog {
	entry := &widget.Entry{PlaceHolder: TR.Trans("label.set_default_smdp_entry_placeholder")}
	form := []*widget.FormItem{
		{Text: TR.Trans("label.default_smdp"), Widget: entry},
	}
	d := dialog.NewForm(TR.Trans("label.set_default_smdp_form"), TR.Trans("dialog.submit"), TR.Trans("dialog.cancel"), form, func(b bool) {
		if b {
			if err := LpacChipDefaultSmdp(entry.Text); err != nil {
				ShowLpacErrDialog(err)
			}
			err := RefreshChipInfo()
			if err != nil {
				ShowLpacErrDialog(err)
			}
		}
	}, WMain)
	d.Resize(fyne.Size{
		Width:  510,
		Height: 200,
	})
	return d
}

func ShowLpacErrDialog(err error) {
	go func() {
		l := &widget.Label{Text: fmt.Sprintf("%v", err)}
		content := container.NewVBox(
			container.NewCenter(container.NewHBox(
				widget.NewIcon(theme.ErrorIcon()),
				widget.NewLabel(TR.Trans("dialog.lpac_error")))),
			container.NewCenter(l),
			container.NewCenter(widget.NewLabel(TR.Trans("message.lpac_error"))))
		dialog.ShowCustom(TR.Trans("dialog.error"), TR.Trans("dialog.ok"), content, WMain)
	}()
}

func ShowSelectItemDialog() {
	go func() {
		d := dialog.NewInformation(TR.Trans("dialog.info"), TR.Trans("message.select_item"), WMain)
		d.Show()
	}()
}

func ShowSelectCardReaderDialog() {
	go func() {
		dialog.ShowInformation(TR.Trans("dialog.info"), TR.Trans("message.select_card_reader"), WMain)
	}()
}

func ShowRefreshNeededDialog() {
	go func() {
		dialog.ShowInformation(TR.Trans("dialog.info"), TR.Trans("message.refresh_required")+"\n", WMain)
	}()
}

// InitAidListDialog 创建AID列表选择对话框
func InitAidListDialog(aidEntry *widget.Entry) dialog.Dialog {
	// 搜索输入框
	searchEntry := &widget.Entry{
		PlaceHolder: TR.Trans("label.aid_search_placeholder"),
	}
	
	// 搜索结果列表
	var filteredList []*AidItem
	var selectedAidIndex int = -1
	
	// 初始化：显示所有AID或eUICC相关AID
	if len(AidList) == 0 {
		// 如果AID列表为空，尝试重新加载
		if aidList, err := LoadAidList(); err == nil {
			AidList = aidList
		}
	}
	
	// 默认显示eUICC相关的AID
	filteredList = make([]*AidItem, 0)
	for _, item := range AidList {
		if item.IsEuicc {
			filteredList = append(filteredList, item)
		}
	}
	// 如果没有eUICC相关的，显示所有
	if len(filteredList) == 0 {
		filteredList = AidList
	}
	
	// 自动判断最佳AID
	bestAid := FindBestAid(ConfigInstance.LpacAID, filteredList)
	if bestAid != nil {
		// 找到最佳AID在列表中的位置
		for i, item := range filteredList {
			if item.AID == bestAid.AID {
				selectedAidIndex = i
				break
			}
		}
	}
	
	// 列表组件
	list := &widget.List{
		Length: func() int {
			return len(filteredList)
		},
		CreateItem: func() fyne.CanvasObject {
			aidLabel := &widget.Label{TextStyle: fyne.TextStyle{Monospace: true}}
			descLabel := &widget.Label{}
			euiccLabel := &widget.Label{TextStyle: fyne.TextStyle{Bold: true}}
			return container.NewVBox(
				container.NewHBox(aidLabel, layout.NewSpacer(), euiccLabel),
				descLabel,
			)
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(filteredList) {
				return
			}
			item := filteredList[i]
			container := o.(*fyne.Container)
			row1 := container.Objects[0].(*fyne.Container)
			aidLabel := row1.Objects[0].(*widget.Label)
			euiccLabel := row1.Objects[2].(*widget.Label)
			descLabel := container.Objects[1].(*widget.Label)
			
			aidLabel.SetText(item.AID)
			descLabel.SetText(item.Description)
			
			if item.IsEuicc {
				euiccLabel.SetText(TR.Trans("label.aid_euicc_tag"))
				euiccLabel.Show()
			} else {
				euiccLabel.Hide()
			}
			
			// 高亮选中的项
			if i == selectedAidIndex {
				aidLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
			} else {
				aidLabel.TextStyle = fyne.TextStyle{Monospace: true}
			}
		},
		OnSelected: func(id widget.ListItemID) {
			selectedAidIndex = id
		},
		OnUnselected: func(id widget.ListItemID) {
			selectedAidIndex = -1
		},
	}
	
	// 搜索功能
	searchEntry.OnChanged = func(query string) {
		filteredList = SearchAidList(query, AidList)
		selectedAidIndex = -1
		list.Refresh()
	}
	
	// 选择按钮
	selectButton := &widget.Button{
		Text: TR.Trans("label.aid_select_button"),
		Icon: theme.ConfirmIcon(),
		OnTapped: func() {
			if selectedAidIndex >= 0 && selectedAidIndex < len(filteredList) {
				aidEntry.SetText(filteredList[selectedAidIndex].AID)
			}
		},
	}
	
	// 自动选择最佳AID按钮（基于当前配置）
	autoSelectButton := &widget.Button{
		Text: TR.Trans("label.aid_auto_select_button"),
		Icon: theme.SearchIcon(),
		OnTapped: func() {
			bestAid := FindBestAid(ConfigInstance.LpacAID, filteredList)
			if bestAid != nil {
				aidEntry.SetText(bestAid.AID)
			}
		},
	}
	
	// 自动测试AID按钮（实际测试哪个能读取卡片）
	testAidButton := &widget.Button{
		Text: TR.Trans("label.aid_test_button"),
		Icon: theme.ConfirmIcon(),
		OnTapped: func() {
			// 检查是否已选择读卡器
			if ConfigInstance.DriverIFID == "" {
				dialog.ShowInformation(TR.Trans("dialog.info"), TR.Trans("message.select_card_reader"), WMain)
				return
			}
			
			// 显示测试进度对话框
			progressLabel := widget.NewLabel(TR.Trans("message.aid_testing"))
			progressDialog := dialog.NewCustomWithoutButtons(TR.Trans("dialog.aid_test"), 
				container.NewVBox(progressLabel), WMain)
			progressDialog.Resize(fyne.Size{Width: 400, Height: 150})
			progressDialog.Show()
			
			// 在goroutine中执行测试
			go func() {
				defer progressDialog.Hide()
				
				// 获取要测试的AID列表（优先测试eUICC相关的）
				testList := make([]*AidItem, 0)
				for _, item := range AidList {
					if item.IsEuicc {
						testList = append(testList, item)
					}
				}
				// 如果没有eUICC相关的，测试所有
				if len(testList) == 0 {
					testList = AidList
				}
				
				// 测试每个AID
				for i, item := range testList {
					progressLabel.SetText(fmt.Sprintf(TR.Trans("message.aid_testing_progress"), 
						i+1, len(testList), item.AID))
					progressDialog.Refresh()
					
					if TestAid(item.AID) {
						// 找到有效的AID
						aidEntry.SetText(item.AID)
						dialog.ShowInformation(TR.Trans("dialog.aid_test_success"),
							fmt.Sprintf(TR.Trans("message.aid_test_found"), item.AID, item.Description), WMain)
						return
					}
				}
				
				// 如果eUICC相关的都不行，测试其他AID
				if len(testList) < len(AidList) {
					for _, item := range AidList {
						if !item.IsEuicc {
							progressLabel.SetText(fmt.Sprintf(TR.Trans("message.aid_testing_progress"), 
								len(testList)+1, len(AidList), item.AID))
							progressDialog.Refresh()
							
							if TestAid(item.AID) {
								aidEntry.SetText(item.AID)
								dialog.ShowInformation(TR.Trans("dialog.aid_test_success"),
									fmt.Sprintf(TR.Trans("message.aid_test_found"), item.AID, item.Description), WMain)
								return
							}
						}
					}
				}
				
				// 没有找到有效的AID
				dialog.ShowInformation(TR.Trans("dialog.aid_test_failed"),
					TR.Trans("message.aid_test_not_found"), WMain)
			}()
		},
	}
	
	// 如果当前有选中的项，启用选择按钮
	if selectedAidIndex >= 0 && selectedAidIndex < len(filteredList) {
		selectButton.Enable()
	} else {
		selectButton.Disable()
	}
	
	// 监听列表选择变化
	originalOnSelected := list.OnSelected
	list.OnSelected = func(id widget.ListItemID) {
		originalOnSelected(id)
		if id >= 0 && id < len(filteredList) {
			selectButton.Enable()
		} else {
			selectButton.Disable()
		}
	}
	
	// 取消按钮
	cancelButton := &widget.Button{
		Text: TR.Trans("dialog.cancel"),
		Icon: theme.CancelIcon(),
		OnTapped: func() {},
	}
	
	// 创建对话框
	content := container.NewBorder(
		container.NewVBox(
			&widget.Label{Text: TR.Trans("label.aid_list_dialog_title"), TextStyle: fyne.TextStyle{Bold: true}},
			searchEntry,
		),
		container.NewHBox(
			autoSelectButton,
			testAidButton,
			layout.NewSpacer(),
			cancelButton,
			selectButton,
		),
		nil,
		nil,
		container.NewScroll(list),
	)
	
	d := dialog.NewCustomWithoutButtons(TR.Trans("label.aid_list_dialog_title"), content, WMain)
	
	// 设置对话框大小
	d.Resize(fyne.Size{
		Width:  700,
		Height: 500,
	})
	
	// 双击选择功能（需要在对话框创建后设置）
	list.OnDoubleTapped = func(id widget.ListItemID) {
		if id >= 0 && id < len(filteredList) {
			aidEntry.SetText(filteredList[id].AID)
			d.Hide()
		}
	}
	
	// 取消按钮关闭对话框
	cancelButton.OnTapped = func() {
		d.Hide()
	}
	
	// 选择按钮关闭对话框
	originalSelectTapped := selectButton.OnTapped
	selectButton.OnTapped = func() {
		originalSelectTapped()
		d.Hide()
	}
	
	return d
}
