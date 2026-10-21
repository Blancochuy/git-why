package git

func (r *Repo) GetCommitDiff(hash string) (string, error) {
	output, err := r.runGit("show", "--no-notes", "--format=", hash)
	if err != nil {
		return "", err
	}
	return output, nil
}
