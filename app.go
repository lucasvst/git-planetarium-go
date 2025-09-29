package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// RepositoryInfo armazena informações sobre um repositório Git.
// As tags `json:"..."` são usadas para serialização/desserialização para JSON,
// equivalente ao `#[derive(serde::Serialize)]` do Rust.
type RepositoryInfo struct {
	Name           string `json:"name"`
	LastCommitDate string `json:"last_commit_date"`
	BranchCount    uint32 `json:"branch_count"`
}

// runGitCommand é uma função auxiliar para executar comandos git em um diretório específico.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir // Define o diretório de trabalho para o comando

	output, err := cmd.Output() // Usamos Output() para capturar apenas stdout
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Setup cria um diretório no caminho especificado se ele não existir.
func (a *App) Setup(path string) (string, error) {
	// Verifica se o caminho já existe e é um diretório
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return fmt.Sprintf("Diretório já existe: %s", path), nil
	}

	// Se o caminho não existe, cria o diretório (e todos os pais, se necessário)
	if os.IsNotExist(err) {
		// 0755 é uma permissão padrão para diretórios (leitura/execução para todos, escrita para o dono)
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("erro ao criar o diretório: %w", err)
		}
		return fmt.Sprintf("Diretório criado com sucesso: %s", path), nil
	}

	// Retorna outros erros que possam ter ocorrido (ex: o caminho é um arquivo)
	return "", err
}

// GitClone clona um repositório para um diretório de destino.
func (a *App) GitClone(repoURL, targetDir string) (string, error) {
	// Equivalente ao Command::new("git").arg("clone")... do Rust
	cmd := exec.Command("git", "clone", repoURL, targetDir)

	// CombinedOutput executa o comando e retorna a saída combinada (stdout e stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Se houver um erro, a saída geralmente contém a mensagem de erro do git
		return "", fmt.Errorf("erro ao clonar repositório: %s\n%w", string(output), err)
	}

	return fmt.Sprintf("Repositório clonado com sucesso em: %s", targetDir), nil
}

// ListRepositories lista todos os repositórios Git dentro de um diretório pai.
func (a *App) ListRepositories(path string) ([]RepositoryInfo, error) {
	// Verifica se o caminho é um diretório válido
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("o caminho não é um diretório válido: %s", path)
	}

	// Em Go, um slice dinâmico é o equivalente ao Vec do Rust.
	var repos []RepositoryInfo

	// os.ReadDir é a forma idiomática de ler as entradas de um diretório.
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o diretório: %w", err)
	}

	for _, entry := range entries {
		// Considera apenas as subpastas
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(path, entry.Name())
		gitPath := filepath.Join(entryPath, ".git")

		// Verifica se a pasta .git existe para confirmar que é um repositório
		if gitInfo, err := os.Stat(gitPath); err == nil && gitInfo.IsDir() {
			// É um repositório Git, vamos coletar as informações

			// Comando para pegar a data do último commit
			lastCommitDate, err := runGitCommand(entryPath, "log", "-1", "--format=%cd")
			if err != nil {
				// Se o comando falhar, definimos um valor padrão, assim como no Rust
				lastCommitDate = "N/A"
			}

			// Comando para contar o número de branches
			branchOutput, err := runGitCommand(entryPath, "branch", "-a")
			var branchCount uint32
			if err != nil {
				branchCount = 0 // Valor padrão em caso de erro
			} else {
				// Conta as linhas não vazias na saída
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
