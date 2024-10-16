package database

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSyncFileBy(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("get-sync-file-by.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))

	user, err := CreateUser(testdb, "test-user", "test-user@example.com", "not a secure password")
	assert.NoError(t, err)
	syncfiles := []*SyncFile{
		{Filepath: "/folder/file1.md", Etag: "f0f9ef0cbb7e0d836aea4a4c6fe6420a"},
		{Filepath: "/folder/file2.md", Etag: "8c42bf48c4b5d8553ad3ab5b30b484df"},
		{Filepath: "/folder/file3.md", Etag: "37e904b58a2a5e61babc827ded3a828d"},
	}
	for i, syncfile := range syncfiles {
		syncfiles[i], err = CreateSyncFile(testdb, syncfile.Filepath, syncfile.Etag, user.Id)
		assert.NoError(t, err)
	}

	testCases := []struct {
		name         string
		id           int64
		userId       int64
		filepath     string
		wantSyncfile *SyncFile
		wantErr      error
	}{
		{
			name:         "found by id",
			id:           int64(syncfiles[1].Id),
			wantSyncfile: syncfiles[1],
		},
		{
			name:    "not found by id",
			id:      int64(syncfiles[2].Id + 1),
			wantErr: ErrNoResults,
		},
		{
			name:         "found by filepath",
			filepath:     "/folder/file1.md",
			wantSyncfile: syncfiles[0],
		},
		{
			name:     "not found by filepath",
			filepath: "/folder/file4.md",
			wantErr:  ErrNoResults,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			var (
				syncfile *SyncFile
				err      error
			)

			if tc.id > 0 {
				syncfile, err = GetSyncFileById(testdb, uint64(tc.id))
			}
			if len(tc.filepath) > 0 {
				syncfile, err = GetSyncFileByFilepath(testdb, tc.filepath)
			}

			assert.ErrorIs(t, err, tc.wantErr)
			if tc.wantSyncfile != nil && tc.wantErr == nil {
				assert.Equal(t, *tc.wantSyncfile, *syncfile)
			}
		})
	}

	// sync files for user are found
	dbSyncfiles, err := GetSyncFilesByUserId(testdb, user.Id)
	assert.NoError(t, err)
	slices.SortFunc(dbSyncfiles, func(a *SyncFile, b *SyncFile) int {
		return int(a.Id - b.Id)
	})
	assert.ElementsMatch(t, syncfiles, dbSyncfiles)

	// sync files for user are not found
	dbSyncfiles, err = GetSyncFilesByUserId(testdb, user.Id+1)
	assert.ErrorIs(t, err, ErrNoResults)
	assert.Empty(t, dbSyncfiles)

	for _, syncfile := range syncfiles {
		_, err := testdb.Exec("DELETE FROM file_syncs WHERE id=?", syncfile.Id)
		assert.NoError(t, err)
	}
	_, err = testdb.Exec("DELETE FROM users WHERE id=?", user.Id)
	assert.NoError(t, err)
}

func TestUpdateFileSyncFilepath(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-update-file-sync-filepath.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))
	user, err := CreateUser(testdb, "test-user", "test-user@example.com", "npt a password")
	assert.NoError(t, err)
	syncfiles := []*SyncFile{
		{Filepath: "/folder/file1.md", Etag: "f0f9ef0cbb7e0d836aea4a4c6fe6420a"},
		{Filepath: "/folder/file2.md", Etag: "8c42bf48c4b5d8553ad3ab5b30b484df"},
		{Filepath: "/folder/file3.md", Etag: "37e904b58a2a5e61babc827ded3a828d"},
	}
	for i, syncfile := range syncfiles {
		syncfiles[i], err = CreateSyncFile(testdb, syncfile.Filepath, syncfile.Etag, user.Id)
		assert.NoError(t, err)
	}

	testCases := []struct {
		name         string
		currFilepath string
		newFilepath  string
		wantErr      error
	}{
		{
			name:         "update filepath successful",
			currFilepath: syncfiles[1].Filepath,
			newFilepath:  "folder/file4.md",
		},
		{
			name:         "update filepath results in conflict",
			currFilepath: syncfiles[1].Filepath,
			newFilepath:  syncfiles[0].Filepath,
			wantErr:      ErrFilepathExists,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// update syncfile
			err := UpdateSyncFileFilepath(testdb, tc.currFilepath, tc.newFilepath)
			assert.ErrorIs(t, err, tc.wantErr)
			if err == nil {
				return
			}

			// make sure change occured in the database
			var count int
			row := testdb.QueryRow("SELECT COUNT(*) FROM file_syncs WHERE filepath=?", tc.newFilepath)
			assert.NoError(t, row.Scan(&count))
			assert.Equal(t, 1, count)
		})
	}
}

func TestUpdateFileSyncEtag(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-update-file-sync-etag.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))
	user, err := CreateUser(testdb, "test-user", "test-user@example.com", "npt a password")
	assert.NoError(t, err)
	syncfiles := []*SyncFile{
		{Filepath: "/folder/file1.md", Etag: "f0f9ef0cbb7e0d836aea4a4c6fe6420a"},
		{Filepath: "/folder/file2.md", Etag: "8c42bf48c4b5d8553ad3ab5b30b484df"},
		{Filepath: "/folder/file3.md", Etag: "37e904b58a2a5e61babc827ded3a828d"},
	}
	for i, syncfile := range syncfiles {
		syncfiles[i], err = CreateSyncFile(testdb, syncfile.Filepath, syncfile.Etag, user.Id)
		assert.NoError(t, err)
	}

	var count int
	row := testdb.QueryRow(
		"SELECT COUNT(*) FROM file_syncs WHERE filepath=? AND etag=?",
		syncfiles[1].Filepath,
		syncfiles[1].Etag,
	)
	assert.NoError(t, row.Scan(&count))
	assert.Equal(t, 1, count)
}
