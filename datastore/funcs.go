package datastore

// Todo Model
type FTP struct {
	ProftpdAccessLogID int    `db:"proftpd_access_log_id" json:"proftpd_access_log_id"`
	Sync               bool   `ip:"sync" json:"sync"`
	IP                 string `db:"ip,omitempty" json:"ip,omitempty"`
	Username           string `db:"username,omitempty" json:"username,omitempty"`
	Time               string `db:"time,omitempty" json:"time,omitempty"`
	Operation          string `db:"operation,omitempty" json:"operation,omitempty"`
	Filename           string `db:"filename,omitempty" json:"filename,omitempty"`
}

// GetAllTodo : Fetch all files to be synced
func GetAllFilesToSync() ([]FTP, error) {
	files := []FTP{}
	rows, err := DB.Queryx("SELECT proftpd_access_log_id,sync,ip,username,time,operation,filename FROM proftpd_access_log where sync=false")
	if err != nil {
		return files, err
	}
	defer rows.Close()
	for rows.Next() {
		file := FTP{}
		err := rows.StructScan(&file)
		if err != nil {
			return files, err
		}

		files = append(files, file)
	}
	return files, nil
}
