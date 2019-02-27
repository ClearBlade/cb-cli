package cblib

import (
	//"fmt"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

const BACKUP_DIRECTORY_SUFFIX = "_cb_bak"

type compare func(sliceOfSystemResources *[]interface{}, i int, j int) bool

func setupAddrs(paddr string, maddr string) {
	cb.CB_ADDR = paddr

	preIdx := strings.Index(paddr, "://")
	if maddr == "" {
		if preIdx != -1 {
			maddr = paddr[preIdx+3:]
		} else {
			maddr = paddr
		}
	}
	postIdx := strings.Index(maddr, ":")
	if postIdx != -1 {
		cb.CB_MSG_ADDR = maddr[:postIdx] + ":1883"
	} else {
		cb.CB_MSG_ADDR = maddr + ":1883"
	}
}

func convertPermissionsNames(perms map[string]interface{}) map[string]interface{} {
	rval := map[string]interface{}{}
	for key, val := range perms {
		switch key {
		case "CodeServices":
			rval["services"] = val
		case "Collections":
			rval["collections"] = val
		case "DevicesList":
			rval["devices"] = val
		case "MsgHistory":
			rval["messagehistory"] = val
		case "Portals":
			rval["portals"] = val
		case "Push":
			rval["push"] = val
		case "Topics":
			rval["topics"] = val
		case "UsersList":
			rval["users"] = val
		default:
			rval[key] = "unknown"
		}
	}
	return rval
}

// Bubble sort, compare by map key
func sortByMapKey(arrayPointer *[]interface{}, sortKey string) {
	if arrayPointer == nil {
		return
	}
	array := *arrayPointer
	swapped := true
	for swapped {
		swapped = false
		for i := 0; i < len(array)-1; i++ {
			needToSwap := compareWithKey(sortKey, arrayPointer, i+1, i)
			if needToSwap {
				swap(arrayPointer, i, i+1)
				swapped = true
			}
		}
	}
}

// Bubble sort, compare by function
func sortByFunction(arrayPointer *[]interface{}, compareFn compare) {
	if arrayPointer == nil {
		return
	}
	array := *arrayPointer
	swapped := true
	for swapped {
		swapped = false
		for i := 0; i < len(array)-1; i++ {
			needToSwap := compareFn(arrayPointer, i+1, i)
			if needToSwap {
				swap(arrayPointer, i, i+1)
				swapped = true
			}
		}
	}
}

func swap(array *[]interface{}, i, j int) {
	tmp := (*array)[j]
	(*array)[j] = (*array)[i]
	(*array)[i] = tmp
}

func isString(input interface{}) bool {
	return input != nil && reflect.TypeOf(input).Name() == "string"
}

func compareWithKey(sortKey string, sliceOfCodeServices *[]interface{}, i, j int) bool {
	slice := *sliceOfCodeServices

	map1, castSuccess1 := slice[i].(map[string]interface{})
	map2, castSuccess2 := slice[j].(map[string]interface{})

	if !castSuccess1 || !castSuccess2 {
		return false
	}

	name1 := map1[sortKey]
	name2 := map2[sortKey]

	if !isString(name1) || !isString(name2) {
		return false
	}

	return name1.(string) < name2.(string)
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createFilePath(args ...string) string {
	return strings.Join(args, "/")
}

func copyMap(daMap map[string]interface{}) map[string]interface{} {
	rtn := make(map[string]interface{})
	for k, v := range daMap {
		rtn[k] = v
	}
	return rtn
}

func getBackupDirectoryName(directoryName string) string {
	return directoryName + BACKUP_DIRECTORY_SUFFIX
}

func removeBackupDirectory(directoryName string) error {
	return removeDirectory(getBackupDirectoryName(directoryName))
}

func backupAndCleanDirectory(directoryName string) error {
	if err := backupDirectory(directoryName); err != nil {
		return err
	}
	return removeDirectoryContents(directoryName)
}

func removeDirectoryContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func backupDirectory(directoryName string) error {
	return copyDir(directoryName, getBackupDirectoryName(directoryName))
}

func removeDirectory(directoryName string) error {
	if err := os.RemoveAll(directoryName); err != nil && err != os.ErrNotExist {
		// if we have an error that doesn't relate to the directory not existing, let's return the error
		return err
	}
	return nil
}

func restoreBackupDirectory(directoryName string) error {
	if err := removeDirectory(directoryName); err != nil && err != os.ErrNotExist {
		fmt.Printf("Error while restoring backup directory for '%s'; Unable to remove destination directory", directoryName)
		return err
	}
	if err := copyDir(getBackupDirectoryName(directoryName), directoryName); err != nil {
		fmt.Printf("Error while restoring backup directory for '%s'; Unable to copy backup directory", directoryName)
		return err
	}
	if err := removeDirectory(getBackupDirectoryName(directoryName)); err != nil {
		fmt.Printf("Error while restoring backup directory for '%s'; Unable to remove backup directory", directoryName)
		return err
	}
	return nil
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func IsInRepo() bool {
	return FoundSystemDotJSON()
}

func FoundSystemDotJSON() bool {
	if _, err := getDict("system.json"); err == nil {
		return true
	}
	return false

}

func FoundCBMeta() bool {
	if _, err := getDict(".cbmeta"); err == nil {
		return true
	}
	return false

}

// These keys are generated upon GET, and not representative of the data model
// If we store to filesystem with these keys, the corresponding PUT/POST for portal fails
func removeBlacklistedPortalKeys(portal map[string]interface{}) map[string]interface{} {
	var blacklist = []string{"permissions", "plugins"}
	for _, key := range blacklist {
		delete(portal, key)
	}
	return portal
}

type ColumnDiff struct {
	add    []map[string]interface{}
	remove []map[string]interface{}
}

func isDefaultColumn(defaultColumns []string, colName string) bool {
	for i := 0; i < len(defaultColumns); i++ {
		if defaultColumns[i] == colName {
			return true
		}
	}
	return false
}

func findDiff(columnListA []map[string]interface{}, columnListB []map[string]interface{}, defaultColumns []string) []map[string]interface{} {
	rtn := make([]map[string]interface{}, 0)
	for i := 0; i < len(columnListA); i++ {
		colName := columnListA[i]["ColumnName"].(string)
		// skip the column if it's a default
		if !isDefaultColumn(defaultColumns, colName) {
			found := false
			for j := 0; j < len(columnListB); j++ {
				if colName == columnListB[j]["ColumnName"].(string) {
					found = true
					break
				}
			}
			if !found {
				rtn = append(rtn, columnListA[i])
			}
		}

	}
	return rtn
}

func compareColumns(localColumnSchema []map[string]interface{}, backendColumnSchema []map[string]interface{}, defaultColumns []string) ColumnDiff {
	diff := ColumnDiff{
		add:    findDiff(localColumnSchema, backendColumnSchema, defaultColumns),
		remove: findDiff(backendColumnSchema, localColumnSchema, defaultColumns),
	}
	return diff
}
