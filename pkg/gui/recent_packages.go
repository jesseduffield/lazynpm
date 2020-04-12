package gui

import (
	"os"

	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func (gui *Gui) mutateRecentPackages(f func([]string) ([]string, bool)) error {
	recentPackages := gui.Config.GetAppState().RecentPackages

	recentPackages, changed := f(recentPackages)
	if !changed {
		return nil
	}

	gui.Config.GetAppState().RecentPackages = recentPackages
	return gui.Config.SaveAppState()

}

func (gui *Gui) sendPackageToTop(path string) error {
	// in case we're not already there, chdir to path
	if err := os.Chdir(path); err != nil {
		return err
	}

	return gui.mutateRecentPackages(func(recentPackages []string) ([]string, bool) {
		updatedRecentPackages := newRecentPackagesList(recentPackages, path)
		// just unconditionally saying we updated it even if we didn't
		return updatedRecentPackages, true
	})
}

func (gui *Gui) removePackage(path string) error {
	return gui.mutateRecentPackages(func(recentPackages []string) ([]string, bool) {
		index, ok := utils.StringIndex(recentPackages, path)
		if !ok {
			// not removing it if it's already been removed
			return nil, false
		}
		updatedRecentPackages := append(recentPackages[:index], recentPackages[index+1:]...)
		return updatedRecentPackages, true
	})
}

func (gui *Gui) addPackage(path string) error {
	return gui.mutateRecentPackages(func(recentPackages []string) ([]string, bool) {
		_, ok := utils.StringIndex(recentPackages, path)
		if ok {
			// not adding it if it's already present
			return nil, false
		}
		updatedRecentPackages := append(recentPackages, path)
		return updatedRecentPackages, true
	})
}

// newRecentPackagesList returns a new repo list with a new entry but only when it doesn't exist yet
// if it already exists, it will be moved to the start of the array
func newRecentPackagesList(recentPackages []string, currentPackage string) []string {
	newPackages := []string{currentPackage}
	for _, pkg := range recentPackages {
		if pkg != currentPackage {
			newPackages = append(newPackages, pkg)
		}
	}
	return newPackages
}
