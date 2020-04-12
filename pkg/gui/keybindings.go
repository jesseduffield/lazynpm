package gui

import (
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

// Binding - a keybinding mapping a key and modifier to a handler. The keypress
// is only handled if the given view has focus, or handled globally if the view
// is ""
type Binding struct {
	ViewName    string
	Contexts    []string
	Handler     func(*gocui.Gui, *gocui.View) error
	Key         interface{} // FIXME: find out how to get `gocui.Key | rune`
	Modifier    gocui.Modifier
	Description string
	Alternative string
}

// GetDisplayStrings returns the display string of a file
func (b *Binding) GetDisplayStrings(isFocused bool) []string {
	return []string{GetKeyDisplay(b.Key), b.Description}
}

var keyMapReversed = map[gocui.Key]string{
	gocui.KeyF1:         "f1",
	gocui.KeyF2:         "f2",
	gocui.KeyF3:         "f3",
	gocui.KeyF4:         "f4",
	gocui.KeyF5:         "f5",
	gocui.KeyF6:         "f6",
	gocui.KeyF7:         "f7",
	gocui.KeyF8:         "f8",
	gocui.KeyF9:         "f9",
	gocui.KeyF10:        "f10",
	gocui.KeyF11:        "f11",
	gocui.KeyF12:        "f12",
	gocui.KeyInsert:     "insert",
	gocui.KeyDelete:     "delete",
	gocui.KeyHome:       "home",
	gocui.KeyEnd:        "end",
	gocui.KeyPgup:       "pgup",
	gocui.KeyPgdn:       "pgdown",
	gocui.KeyArrowUp:    "▲",
	gocui.KeyArrowDown:  "▼",
	gocui.KeyArrowLeft:  "◄",
	gocui.KeyArrowRight: "►",
	gocui.KeyTab:        "tab",        // ctrl+i
	gocui.KeyEnter:      "enter",      // ctrl+m
	gocui.KeyEsc:        "esc",        // ctrl+[, ctrl+3
	gocui.KeyBackspace:  "backspace",  // ctrl+h
	gocui.KeyCtrlSpace:  "ctrl+space", // ctrl+~, ctrl+2
	gocui.KeyCtrlSlash:  "ctrl+/",     // ctrl+_
	gocui.KeySpace:      "space",
	gocui.KeyCtrlA:      "ctrl+a",
	gocui.KeyCtrlB:      "ctrl+b",
	gocui.KeyCtrlC:      "ctrl+c",
	gocui.KeyCtrlD:      "ctrl+d",
	gocui.KeyCtrlE:      "ctrl+e",
	gocui.KeyCtrlF:      "ctrl+f",
	gocui.KeyCtrlG:      "ctrl+g",
	gocui.KeyCtrlJ:      "ctrl+j",
	gocui.KeyCtrlK:      "ctrl+k",
	gocui.KeyCtrlL:      "ctrl+l",
	gocui.KeyCtrlN:      "ctrl+n",
	gocui.KeyCtrlO:      "ctrl+o",
	gocui.KeyCtrlP:      "ctrl+p",
	gocui.KeyCtrlQ:      "ctrl+q",
	gocui.KeyCtrlR:      "ctrl+r",
	gocui.KeyCtrlS:      "ctrl+s",
	gocui.KeyCtrlT:      "ctrl+t",
	gocui.KeyCtrlU:      "ctrl+u",
	gocui.KeyCtrlV:      "ctrl+v",
	gocui.KeyCtrlW:      "ctrl+w",
	gocui.KeyCtrlX:      "ctrl+x",
	gocui.KeyCtrlY:      "ctrl+y",
	gocui.KeyCtrlZ:      "ctrl+z",
	gocui.KeyCtrl4:      "ctrl+4", // ctrl+\
	gocui.KeyCtrl5:      "ctrl+5", // ctrl+]
	gocui.KeyCtrl6:      "ctrl+6",
	gocui.KeyCtrl8:      "ctrl+8",
}

var keymap = map[string]interface{}{
	"<c-a>":       gocui.KeyCtrlA,
	"<c-b>":       gocui.KeyCtrlB,
	"<c-c>":       gocui.KeyCtrlC,
	"<c-d>":       gocui.KeyCtrlD,
	"<c-e>":       gocui.KeyCtrlE,
	"<c-f>":       gocui.KeyCtrlF,
	"<c-g>":       gocui.KeyCtrlG,
	"<c-h>":       gocui.KeyCtrlH,
	"<c-i>":       gocui.KeyCtrlI,
	"<c-j>":       gocui.KeyCtrlJ,
	"<c-k>":       gocui.KeyCtrlK,
	"<c-l>":       gocui.KeyCtrlL,
	"<c-m>":       gocui.KeyCtrlM,
	"<c-n>":       gocui.KeyCtrlN,
	"<c-o>":       gocui.KeyCtrlO,
	"<c-p>":       gocui.KeyCtrlP,
	"<c-q>":       gocui.KeyCtrlQ,
	"<c-r>":       gocui.KeyCtrlR,
	"<c-s>":       gocui.KeyCtrlS,
	"<c-t>":       gocui.KeyCtrlT,
	"<c-u>":       gocui.KeyCtrlU,
	"<c-v>":       gocui.KeyCtrlV,
	"<c-w>":       gocui.KeyCtrlW,
	"<c-x>":       gocui.KeyCtrlX,
	"<c-y>":       gocui.KeyCtrlY,
	"<c-z>":       gocui.KeyCtrlZ,
	"<c-~>":       gocui.KeyCtrlTilde,
	"<c-2>":       gocui.KeyCtrl2,
	"<c-3>":       gocui.KeyCtrl3,
	"<c-4>":       gocui.KeyCtrl4,
	"<c-5>":       gocui.KeyCtrl5,
	"<c-6>":       gocui.KeyCtrl6,
	"<c-7>":       gocui.KeyCtrl7,
	"<c-8>":       gocui.KeyCtrl8,
	"<c-space>":   gocui.KeyCtrlSpace,
	"<c-\\>":      gocui.KeyCtrlBackslash,
	"<c-[>":       gocui.KeyCtrlLsqBracket,
	"<c-]>":       gocui.KeyCtrlRsqBracket,
	"<c-/>":       gocui.KeyCtrlSlash,
	"<c-_>":       gocui.KeyCtrlUnderscore,
	"<backspace>": gocui.KeyBackspace,
	"<tab>":       gocui.KeyTab,
	"<enter>":     gocui.KeyEnter,
	"<esc>":       gocui.KeyEsc,
	"<space>":     gocui.KeySpace,
	"<f1>":        gocui.KeyF1,
	"<f2>":        gocui.KeyF2,
	"<f3>":        gocui.KeyF3,
	"<f4>":        gocui.KeyF4,
	"<f5>":        gocui.KeyF5,
	"<f6>":        gocui.KeyF6,
	"<f7>":        gocui.KeyF7,
	"<f8>":        gocui.KeyF8,
	"<f9>":        gocui.KeyF9,
	"<f10>":       gocui.KeyF10,
	"<f11>":       gocui.KeyF11,
	"<f12>":       gocui.KeyF12,
	"<insert>":    gocui.KeyInsert,
	"<delete>":    gocui.KeyDelete,
	"<home>":      gocui.KeyHome,
	"<end>":       gocui.KeyEnd,
	"<pgup>":      gocui.KeyPgup,
	"<pgdown>":    gocui.KeyPgdn,
	"<up>":        gocui.KeyArrowUp,
	"<down>":      gocui.KeyArrowDown,
	"<left>":      gocui.KeyArrowLeft,
	"<right>":     gocui.KeyArrowRight,
}

func (gui *Gui) getKeyDisplay(name string) string {
	key := gui.getKey(name)
	return GetKeyDisplay(key)
}

func GetKeyDisplay(key interface{}) string {
	keyInt := 0

	switch key := key.(type) {
	case rune:
		keyInt = int(key)
	case gocui.Key:
		value, ok := keyMapReversed[key]
		if ok {
			return value
		}
		keyInt = int(key)
	}

	return string(keyInt)
}

func (gui *Gui) getKey(name string) interface{} {
	key := gui.Config.GetUserConfig().GetString("keybinding." + name)
	if len(key) > 1 {
		binding := keymap[strings.ToLower(key)]
		if binding == nil {
			log.Fatalf("Unrecognized key %s for keybinding %s", strings.ToLower(key), name)
		} else {
			return binding
		}
	} else if len(key) == 1 {
		return []rune(key)[0]
	}
	log.Fatal("Key empty for keybinding: " + strings.ToLower(name))
	return nil
}

// GetInitialKeybindings is a function.
func (gui *Gui) GetInitialKeybindings() []*Binding {
	bindings := []*Binding{
		{
			ViewName: "",
			Key:      gui.getKey("universal.quit"),
			Handler:  gui.handleQuit,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.quitWithoutChangingDirectory"),
			Handler:  gui.handleQuitWithoutChangingDirectory,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.quit-alt1"),
			Handler:  gui.handleQuit,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.return"),
			Handler:  gui.handleQuit,
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.scrollUpMain"),
			Handler:     gui.scrollUpMain,
			Alternative: "fn+up",
			Description: gui.Tr.SLocalize("scrollUpMainPanel"),
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.scrollDownMain"),
			Handler:     gui.scrollDownMain,
			Alternative: "fn+down",
			Description: gui.Tr.SLocalize("scrollDownMainPanel"),
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.scrollUpMain-alt1"),
			Handler:  gui.scrollUpMain,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.scrollDownMain-alt1"),
			Handler:  gui.scrollDownMain,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.scrollUpMain-alt2"),
			Handler:  gui.scrollUpMain,
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.scrollDownMain-alt2"),
			Handler:  gui.scrollDownMain,
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.refresh"),
			Handler:     gui.handleRefresh,
			Description: gui.Tr.SLocalize("refresh"),
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.optionMenu"),
			Handler:     gui.handleCreateOptionsMenu,
			Description: gui.Tr.SLocalize("openMenu"),
		},
		{
			ViewName: "",
			Key:      gui.getKey("universal.optionMenu-alt1"),
			Handler:  gui.handleCreateOptionsMenu,
		},
		{
			ViewName: "",
			Key:      gocui.MouseMiddle,
			Handler:  gui.handleCreateOptionsMenu,
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.kill"),
			Handler:     gui.wrappedHandler(gui.handleKillCommand),
			Description: "kill running command",
		},
		{
			ViewName:    "status",
			Key:         gui.getKey("universal.edit"),
			Handler:     gui.handleEditConfig,
			Description: gui.Tr.SLocalize("EditConfig"),
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.nextScreenMode"),
			Handler:     gui.nextScreenMode,
			Description: gui.Tr.SLocalize("nextScreenMode"),
		},
		{
			ViewName:    "",
			Key:         gui.getKey("universal.prevScreenMode"),
			Handler:     gui.prevScreenMode,
			Description: gui.Tr.SLocalize("prevScreenMode"),
		},
		{
			ViewName:    "status",
			Key:         gui.getKey("universal.openFile"),
			Handler:     gui.handleOpenConfig,
			Description: gui.Tr.SLocalize("OpenConfig"),
		},
		{
			ViewName:    "status",
			Key:         gui.getKey("status.checkForUpdate"),
			Handler:     gui.handleCheckForUpdate,
			Description: gui.Tr.SLocalize("checkForUpdate"),
		},

		{
			ViewName:    "menu",
			Key:         gui.getKey("universal.return"),
			Handler:     gui.handleMenuClose,
			Description: gui.Tr.SLocalize("closeMenu"),
		},
		{
			ViewName:    "menu",
			Key:         gui.getKey("universal.quit"),
			Handler:     gui.handleMenuClose,
			Description: gui.Tr.SLocalize("closeMenu"),
		},
		{
			ViewName: "information",
			Key:      gocui.MouseLeft,
			Handler:  gui.handleInfoClick,
		},
		{
			ViewName: "secondary",
			Contexts: []string{"normal"},
			Key:      gocui.MouseLeft,
			Handler:  gui.handleMouseDownSecondary,
		},
		{
			ViewName: "status",
			Key:      gocui.MouseLeft,
			Handler:  gui.handleStatusClick,
		},
		{
			ViewName: "search",
			Key:      gocui.KeyEnter,
			Handler:  gui.handleSearch,
		},
		{
			ViewName: "search",
			Key:      gui.getKey("universal.return"),
			Handler:  gui.handleSearchEscape,
		},
		{
			ViewName: "confirmation",
			Key:      gui.getKey("universal.prevItem"),
			Handler:  gui.scrollUpConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      gui.getKey("universal.nextItem"),
			Handler:  gui.scrollDownConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      gui.getKey("universal.prevItem-alt"),
			Handler:  gui.scrollUpConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      gui.getKey("universal.nextItem-alt"),
			Handler:  gui.scrollDownConfirmationPanel,
		},
		{
			ViewName: "packages",
			Key:      gui.getKey("universal.select"),
			Handler:  gui.wrappedPackageHandler(gui.handleCheckoutPackage),
		},
		{
			ViewName: "packages",
			Key:      gui.getKey("packages.publish"),
			Handler:  gui.wrappedPackageHandler(gui.handlePublishPackage),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("universal.new"),
			Handler:     gui.wrappedHandler(gui.handleAddPackage),
			Description: "add package to list",
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("packages.pack"),
			Handler:     gui.wrappedPackageHandler(gui.handlePackPackage),
			Description: fmt.Sprintf("%s package", utils.ColoredString("`npm pack`", color.FgYellow)),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("packages.link"),
			Handler:     gui.wrappedHandler(gui.handleLinkPackage),
			Description: fmt.Sprintf("%s (or unlink if already linked)", utils.ColoredString("`npm link <package>`", color.FgYellow)),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("packages.globalLink"),
			Handler:     gui.wrappedPackageHandler(gui.handleGlobalLinkPackage),
			Description: fmt.Sprintf("%s (i.e. globally link) (or unlink if already linked)", utils.ColoredString("`npm link`", color.FgYellow)),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("universal.remove"),
			Handler:     gui.wrappedPackageHandler(gui.handleRemovePackage),
			Description: "remove package from list",
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("universal.install"),
			Handler:     gui.wrappedPackageHandler(gui.handleInstall),
			Description: fmt.Sprintf("%s package", utils.ColoredString("`npm install`", color.FgYellow)),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("packages.build"),
			Handler:     gui.wrappedPackageHandler(gui.handleBuild),
			Description: fmt.Sprintf("%s package", utils.ColoredString("`npm run build`", color.FgYellow)),
		},
		{
			ViewName:    "packages",
			Key:         gui.getKey("universal.openFile"),
			Handler:     gui.wrappedPackageHandler(gui.handleOpenPackageConfig),
			Description: "open package.json",
		},
		{
			ViewName:    "scripts",
			Key:         gui.getKey("universal.select"),
			Handler:     gui.wrappedScriptHandler(gui.handleRunScript),
			Description: fmt.Sprintf("%s script", utils.ColoredString("`npm run`", color.FgYellow)),
		},
		{
			ViewName:    "scripts",
			Key:         gui.getKey("universal.remove"),
			Handler:     gui.wrappedScriptHandler(gui.handleRemoveScript),
			Description: "remove script from package.json",
		},
		{
			ViewName:    "scripts",
			Key:         gui.getKey("universal.edit"),
			Handler:     gui.wrappedScriptHandler(gui.handleEditScript),
			Description: "edit script",
		},
		{
			ViewName:    "scripts",
			Key:         gui.getKey("universal.new"),
			Handler:     gui.wrappedHandler(gui.handleAddScript),
			Description: "add script",
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.install"),
			Handler:     gui.wrappedDependencyHandler(gui.handleDepInstall),
			Description: fmt.Sprintf("%s dependency", utils.ColoredString("`npm install`", color.FgYellow)),
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.openFile"),
			Handler:     gui.wrappedDependencyHandler(gui.handleOpenDepPackageConfig),
			Description: "open package.json",
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.update"),
			Handler:     gui.wrappedDependencyHandler(gui.handleDepUpdate),
			Description: fmt.Sprintf("%s dependency", utils.ColoredString("`npm update`", color.FgYellow)),
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.remove"),
			Handler:     gui.wrappedDependencyHandler(gui.handleDepUninstall),
			Description: fmt.Sprintf("%s dependency", utils.ColoredString("`npm uninstall`", color.FgYellow)),
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("dependencies.changeType"),
			Handler:     gui.wrappedDependencyHandler(gui.handleChangeDepType),
			Description: "change dependency type (prod/dev/optional)",
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.new"),
			Handler:     gui.wrappedDependencyHandler(gui.handleAddDependency),
			Description: fmt.Sprintf("%s new dependency", utils.ColoredString("`npm install`", color.FgYellow)),
		},
		{
			ViewName:    "deps",
			Key:         gui.getKey("universal.edit"),
			Handler:     gui.wrappedDependencyHandler(gui.handleEditDepConstraint),
			Description: "edit dependency constraint",
		},
	}

	for _, viewName := range []string{"status", "packages", "deps", "scripts", "menu"} {
		bindings = append(bindings, []*Binding{
			{ViewName: viewName, Key: gui.getKey("universal.togglePanel"), Handler: gui.nextView},
			{ViewName: viewName, Key: gui.getKey("universal.prevBlock"), Handler: gui.previousView},
			{ViewName: viewName, Key: gui.getKey("universal.nextBlock"), Handler: gui.nextView},
			{ViewName: viewName, Key: gui.getKey("universal.prevBlock-alt"), Handler: gui.previousView},
			{ViewName: viewName, Key: gui.getKey("universal.nextBlock-alt"), Handler: gui.nextView},
		}...)
	}

	// Appends keybindings to jump to a particular sideView using numbers
	for i, viewName := range []string{"status", "packages", "deps", "scripts"} {
		bindings = append(bindings, &Binding{ViewName: "", Key: rune(i+1) + '0', Handler: gui.goToSideView(viewName)})
	}

	for _, listView := range gui.getListViews() {
		bindings = append(bindings, []*Binding{
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.prevItem-alt"), Handler: listView.handlePrevLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.prevItem"), Handler: listView.handlePrevLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gocui.MouseWheelUp, Handler: listView.handlePrevLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.nextItem-alt"), Handler: listView.handleNextLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.nextItem"), Handler: listView.handleNextLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.prevPage"), Handler: listView.handlePrevPage, Description: gui.Tr.SLocalize("prevPage")},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.nextPage"), Handler: listView.handleNextPage, Description: gui.Tr.SLocalize("nextPage")},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gui.getKey("universal.gotoTop"), Handler: listView.handleGotoTop, Description: gui.Tr.SLocalize("gotoTop")},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gocui.MouseWheelDown, Handler: listView.handleNextLine},
			{ViewName: listView.viewName, Contexts: []string{listView.context}, Key: gocui.MouseLeft, Handler: listView.handleClick},
			{
				ViewName:    listView.viewName,
				Contexts:    []string{listView.context},
				Key:         gui.getKey("universal.startSearch"),
				Handler:     gui.handleOpenSearch,
				Description: gui.Tr.SLocalize("startSearch"),
			},
			{
				ViewName:    listView.viewName,
				Contexts:    []string{listView.context},
				Key:         gui.getKey("universal.gotoBottom"),
				Handler:     listView.handleGotoBottom,
				Description: gui.Tr.SLocalize("gotoBottom"),
			},
		}...)
	}

	return bindings
}

func (gui *Gui) keybindings(g *gocui.Gui) error {
	bindings := gui.GetInitialKeybindings()

	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.ViewName, binding.Contexts, binding.Key, binding.Modifier, binding.Handler); err != nil {
			return err
		}
	}

	tabClickBindings := map[string]func(int) error{
		// none yet
	}

	for viewName, binding := range tabClickBindings {
		if err := g.SetTabClickBinding(viewName, binding); err != nil {
			return err
		}
	}

	return nil
}
