/*

Todo list when making a new translation
- Copy this file and rename it to the language you want to translate to like someLanguage.go
- Change the addEnglish() name to the language you want to translate to like addSomeLanguage()
- change the first function argument of i18nObject.AddMessages( to the language you want to translate to like language.SomeLanguage
- Remove this todo and the about section

*/

package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func addEnglish(i18nObject *i18n.Bundle) error {

	return i18nObject.AddMessages(language.English,
		&i18n.Message{
			ID:    "NotEnoughSpace",
			Other: "Not enough space to render panels",
		}, &i18n.Message{
			ID:    "PackagesTitle",
			Other: "Packages",
		}, &i18n.Message{
			ID:    "DepsTitle",
			Other: "Dependencies",
		}, &i18n.Message{
			ID:    "ScriptsTitle",
			Other: "Scripts",
		}, &i18n.Message{
			ID:    "MainTitle",
			Other: "Main",
		}, &i18n.Message{
			ID:    "NormalTitle",
			Other: "Normal",
		}, &i18n.Message{
			ID:    "StatusTitle",
			Other: "Status",
		}, &i18n.Message{
			ID:    "GlobalTitle",
			Other: "Global",
		}, &i18n.Message{
			ID:    "navigate",
			Other: "navigate",
		}, &i18n.Message{
			ID:    "menu",
			Other: "menu",
		}, &i18n.Message{
			ID:    "execute",
			Other: "execute",
		}, &i18n.Message{
			ID:    "open",
			Other: "open",
		}, &i18n.Message{
			ID:    "delete",
			Other: "delete",
		}, &i18n.Message{
			ID:    "refresh",
			Other: "refresh",
		}, &i18n.Message{
			ID:    "edit",
			Other: "edit",
		}, &i18n.Message{
			ID:    "scroll",
			Other: "scroll",
		}, &i18n.Message{
			ID:    "CloseConfirm",
			Other: "{{.keyBindClose}}: close, {{.keyBindConfirm}}: confirm",
		}, &i18n.Message{
			ID:    "close",
			Other: "close",
		}, &i18n.Message{
			ID:    "Error",
			Other: "Error",
		}, &i18n.Message{
			ID:    "resizingPopupPanel",
			Other: "resizing popup panel",
		}, &i18n.Message{
			ID:    "IssntListOfViews",
			Other: "{{.name}} is not in the list of views",
		}, &i18n.Message{
			ID:    "NoViewMachingNewLineFocusedSwitchStatement",
			Other: "No view matching newLineFocused switch statement",
		}, &i18n.Message{
			ID:    "newFocusedViewIs",
			Other: "new focused view is {{.newFocusedView}}",
		}, &i18n.Message{
			ID:    "OpenConfig",
			Other: "open config file",
		}, &i18n.Message{
			ID:    "EditConfig",
			Other: "edit config file",
		}, &i18n.Message{
			ID:    "checkForUpdate",
			Other: "check for update",
		}, &i18n.Message{
			ID:    "CheckingForUpdates",
			Other: "Checking for updates...",
		}, &i18n.Message{
			ID:    "OnLatestVersionErr",
			Other: "You already have the latest version",
		}, &i18n.Message{
			ID:    "MajorVersionErr",
			Other: "New version ({{.newVersion}}) has non-backwards compatible changes compared to the current version ({{.currentVersion}})",
		}, &i18n.Message{
			ID:    "CouldNotFindBinaryErr",
			Other: "Could not find any binary at {{.url}}",
		}, &i18n.Message{
			ID:    "AnonymousReportingTitle",
			Other: "Help make lazynpm better",
		}, &i18n.Message{
			ID:    "AnonymousReportingPrompt",
			Other: "Would you like to enable anonymous reporting data to help improve lazynpm? (enter/esc)",
		}, &i18n.Message{
			ID:    "ShamelessSelfPromotionTitle",
			Other: "Shameless Self Promotion",
		}, &i18n.Message{
			ID:    "ShamelessSelfPromotionMessage",
			Other: `Thanks for using lazynpm! Github are now matching any donations dollar-for-dollar for the next 12 months, so if you've been tossing up over whether to click the donate link in the bottom right corner, now is the time!`,
		}, &i18n.Message{
			ID:    "editFile",
			Other: `edit file`,
		}, &i18n.Message{
			ID:    "openFile",
			Other: `open file`,
		}, &i18n.Message{
			ID:    "ConfirmQuit",
			Other: `Are you sure you want to quit?`,
		},
		&i18n.Message{
			ID:    "TogglePanel",
			Other: `switch to other panel`,
		}, &i18n.Message{
			ID:    "SearchTitle",
			Other: "Search",
		}, &i18n.Message{
			ID:    "MenuTitle",
			Other: "Menu",
		}, &i18n.Message{
			ID:    "InformationTitle",
			Other: "Information",
		}, &i18n.Message{
			ID:    "SecondaryTitle",
			Other: "Secondary",
		}, &i18n.Message{
			ID:    "Title",
			Other: "Title",
		}, &i18n.Message{
			ID:    "GlobalTitle",
			Other: "Global Keybindings",
		}, &i18n.Message{
			ID:    "ErrorOccurred",
			Other: "An error occurred! Please create an issue at https://github.com/jesseduffield/lazynpm/issues",
		}, &i18n.Message{
			ID:    "NoRoom",
			Other: "Not enough room",
		}, &i18n.Message{
			ID:    "Donate",
			Other: "Donate",
		}, &i18n.Message{
			ID:    "PrevLine",
			Other: "select previous line",
		}, &i18n.Message{
			ID:    "NextLine",
			Other: "select next line",
		}, &i18n.Message{
			ID:    "ScrollDown",
			Other: "scroll down",
		}, &i18n.Message{
			ID:    "ScrollUp",
			Other: "scroll up",
		}, &i18n.Message{
			ID:    "scrollUpMainPanel",
			Other: "scroll up main panel",
		}, &i18n.Message{
			ID:    "scrollDownMainPanel",
			Other: "scroll down main panel",
		}, &i18n.Message{
			ID:    "goBack",
			Other: "go back",
		}, &i18n.Message{
			ID:    "cancel",
			Other: "cancel",
		}, &i18n.Message{
			ID:    "executeCustomCommand",
			Other: "execute custom command",
		}, &i18n.Message{
			ID:    "CustomCommand",
			Other: "Custom Command:",
		}, &i18n.Message{
			ID:    "pressEnterToReturn",
			Other: "Press enter to return to lazynpm",
		}, &i18n.Message{
			ID:    "jump",
			Other: "jump to panel",
		}, &i18n.Message{
			ID:    "nextScreenMode",
			Other: "next screen mode (normal/half/fullscreen)",
		}, &i18n.Message{
			ID:    "prevScreenMode",
			Other: "prev screen mode",
		}, &i18n.Message{
			ID:    "startSearch",
			Other: "start search",
		}, &i18n.Message{
			ID:    "Panel",
			Other: "Panel",
		}, &i18n.Message{
			ID:    "Keybindings",
			Other: "Keybindings",
		}, &i18n.Message{
			ID:    "openMenu",
			Other: "open menu",
		}, &i18n.Message{
			ID:    "closeMenu",
			Other: "close menu",
		}, &i18n.Message{
			ID:    "nextTab",
			Other: "next tab",
		}, &i18n.Message{
			ID:    "prevTab",
			Other: "previous tab",
		}, &i18n.Message{
			ID:    "prevPage",
			Other: "previous page",
		}, &i18n.Message{
			ID:    "nextPage",
			Other: "next page",
		}, &i18n.Message{
			ID:    "gotoTop",
			Other: "scroll to top",
		}, &i18n.Message{
			ID:    "gotoBottom",
			Other: "scroll to bottom",
		}, &i18n.Message{
			ID:    "(reset)",
			Other: "(reset)",
		}, &i18n.Message{
			ID:    "NoDependencies",
			Other: "No dependencies",
		}, &i18n.Message{
			ID:    "NoScripts",
			Other: "No Scripts",
		},
	)
}
