package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type folderRecord struct {
	folderFields

	DeletedAt        *string  `json:"deleted_at"`
	FormerParentPath []string `json:"former_parent_path"`
}

func folderRecordOf(trash domain.Trash) func(folder domain.Folder) folderRecord {
	return func(folder domain.Folder) folderRecord {
		return folderRecord{
			folderFields:     folderFieldsOf(folder),
			DeletedAt:        deletedAtOf(trash.Folders, folder.ID()),
			FormerParentPath: formerPathOf(trash.Folders, folder.ID()),
		}
	}
}

func (r folderRecord) entry() (domain.LibraryEntry[domain.Folder], error) {
	folder, err := r.folder()
	if err != nil {
		return domain.LibraryEntry[domain.Folder]{}, err
	}

	trash, err := parseTrashFields(r.DeletedAt, r.FormerParentPath)
	if err != nil {
		return domain.LibraryEntry[domain.Folder]{}, r.named(err)
	}

	entry, err := domain.NewLibraryEntry(folder, trash.deletedAt, trash.formerPath)
	if err != nil {
		return domain.LibraryEntry[domain.Folder]{}, r.named(err)
	}

	return entry, nil
}
