package git_manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitManager struct {
	ctx context.Context
}

func NewGitManager() *GitManager {
	return &GitManager{}
}

func (a *GitManager) startup(ctx context.Context) {
	a.ctx = ctx
}

type RepositoryInfo struct {
	Name           string `json:"name"`
	LastCommitDate string `json:"last_commit_date"`
	BranchCount    uint32 `json:"branch_count"`
}

func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (a *GitManager) Setup(path string) (string, error) {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return fmt.Sprintf("Diretório já existe: %s", path), nil
	}

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("erro ao criar o diretório: %w", err)
		}
		return fmt.Sprintf("Diretório criado com sucesso: %s", path), nil
	}

	return "", err
}

func (a *GitManager) GitClone(repoURL, targetDir string) (string, error) {
	cmd := exec.Command("git", "clone", repoURL, targetDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao clonar repositório: %s\n%w", string(output), err)
	}

	return fmt.Sprintf("Repositório clonado com sucesso em: %s", targetDir), nil
}

func (a *GitManager) ListRepositories(path string) ([]RepositoryInfo, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("o caminho não é um diretório válido: %s", path)
	}

	var repos []RepositoryInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o diretório: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(path, entry.Name())
		gitPath := filepath.Join(entryPath, ".git")

		if gitInfo, err := os.Stat(gitPath); err == nil && gitInfo.IsDir() {

			lastCommitDate, err := runGitCommand(entryPath, "log", "-1", "--format=%cd")
			if err != nil {
				lastCommitDate = "N/A"
			}

			branchOutput, err := runGitCommand(entryPath, "branch", "-a")
			var branchCount uint32
			if err != nil {
				branchCount = 0
			} else {
				lines := strings.Split(branchOutput, "\n")
				branchCount = uint32(len(lines))
			}

			repos = append(repos, RepositoryInfo{
				Name:           entry.Name(),
				LastCommitDate: lastCommitDate,
				BranchCount:    branchCount,
			})
		}
	}

	return repos, nil
}
