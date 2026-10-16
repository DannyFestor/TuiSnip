package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type folderRecord struct {
	folderRecordV1

	DeletedAt        *string  `json:"deleted_at"`
	FormerParentPath []string `json:"former_parent_path"`
}

func folderRecordOf(trash domain.Trash) func(folder domain.Folder) folderRecord {
	return func(folder domain.Folder) folderRecord {
		return folderRecord{
			folderRecordV1:   folderRecordV1Of(folder),
			DeletedAt:        deletedAtOf(trash.Folders, folder.ID()),
			FormerParentPath: formerPathOf(trash.Folders, folder.ID()),
		}
	}
}

func (r folderRecord) entry() (trashEntry[domain.Folder], error) {
	folder, err := r.folder()
	if err != nil {
		return trashEntry[domain.Folder]{}, err
	}

	entry, err := entryOf(folder, r.DeletedAt, r.FormerParentPath)
	if err != nil {
		return trashEntry[domain.Folder]{}, r.named(err)
	}

	return entry, nil
}
