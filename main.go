package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

var arquivoLog *os.File
var statusSites = make(map[string]bool)
var sitesParaMonitorar []string
var servidorLogIniciado bool

func main() {
	var err error
	arquivoLog, err = os.OpenFile("monitoramento.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo de log:", err)
		return
	}
	defer arquivoLog.Close()

	sitesParaMonitorar = sitesFastShop()

	for {
		exibirIntroducao()
		comando := lerComando()

		switch comando {
		case 1:
			iniciarMonitoramento(sitesParaMonitorar)
		case 2:
			exibirLogs()
		case 3:
			iniciarMonitoramento(sitesFastShop())
		case 4:
			iniciarMonitoramento(sitesMCSab())
		case 5:
			gerenciarSites()
		case 6:
			if servidorLogIniciado {
				color.Yellow("Servidor de logs já está em execução.")
			} else {
				iniciarServidorLog()
				servidorLogIniciado = true
			}
		case 7:
			limparLogs()
		case 8:
			limparTela()
		case 0:
			color.Green("Saindo do Programa. Até mais!")
			os.Exit(0)
		default:
			color.Red("Comando inválido. Tente novamente.")
		}
		exibirRodape()
	}
}

func lerComando() int {
	var comando int
	for {
		fmt.Print("Digite o comando: ")
		_, err := fmt.Scan(&comando)
		if err != nil {
			color.Red("Entrada inválida, tente novamente.")
			bufio.NewReader(os.Stdin).ReadString('\n')
			continue
		}
		return comando
	}
}

func exibirIntroducao() {
	limparTela()
	fmt.Println("\n=========================")
	color.Blue("   Monitoramento de Sites  ")
	fmt.Println("=========================")
	color.Cyan("1 - Iniciar monitoramento dos sites configurados")
	color.Cyan("2 - Exibir logs de monitoramento")
	color.Cyan("3 - Monitorar sites da Fast-Shop")
	color.Cyan("4 - Monitorar sites da MCASSAB")
	color.Cyan("5 - Gerenciar sites de monitoramento")
	color.Cyan("6 - Iniciar servidor para acessar logs")
	color.Cyan("7 - Limpar logs")
	color.Cyan("8 - Limpar tela")
	color.Cyan("0 - Sair do Programa")
	fmt.Println("=========================\n")
}

func iniciarServidorLog() {
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := os.ReadFile("monitoramento.log")
		if err != nil {
			http.Error(w, "Erro ao ler o arquivo de log", http.StatusInternalServerError)
			return
		}
		html := `
		<!DOCTYPE html>
		<html lang="pt-BR">
		<head>
			<meta charset="UTF-8">
			<title>Logs de Monitoramento</title>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f4; color: #333; padding: 20px; }
				h1 { color: #007BFF; }
				.log-entry { margin: 10px 0; padding: 10px; border-radius: 5px; }
				.log-up { background-color: #d4edda; color: #155724; }
				.log-down { background-color: #f8d7da; color: #721c24; }
				.log-error { background-color: #fff3cd; color: #856404; }
				.separator { border-top: 1px solid #ccc; margin: 10px 0; }
			</style>
		</head>
		<body>
			<h1>Logs de Monitoramento</h1>
			<div>`
		lines := bufio.NewScanner(strings.NewReader(string(data)))
		for lines.Scan() {
			line := lines.Text()
			var class string
			if strings.Contains(line, "UP") {
				class = "log-up"
			} else if strings.Contains(line, "DOWN") {
				class = "log-down"
			} else if strings.Contains(line, "Erro") {
				class = "log-error"
			}
			html += `<div class="log-entry ` + class + `">` + line + `</div><div class="separator"></div>`
		}
		if err := lines.Err(); err != nil {
			http.Error(w, "Erro ao ler as linhas do log", http.StatusInternalServerError)
			return
		}
		html += `</div></body></html>`
		fmt.Fprint(w, html)
	})

	go func() {
		fmt.Println("Servidor de logs iniciado em http://localhost:8080/logs")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			color.Red("Erro ao iniciar o servidor: %s", err)
		}
	}()

	time.Sleep(1 * time.Second)
}

func gerenciarSites() {
	for {
		fmt.Println("Gerenciar Sites de Monitoramento")
		color.Yellow("1 - Adicionar um novo site")
		color.Yellow("2 - Remover um site")
		color.Yellow("3 - Listar sites atuais")
		color.Yellow("0 - Voltar ao menu principal")
		fmt.Print("Escolha uma opção: ")
		var opcao int
		_, err := fmt.Scan(&opcao)
		if err != nil {
			color.Red("Entrada inválida, tente novamente.")
			bufio.NewReader(os.Stdin).ReadString('\n')
			continue
		}
		switch opcao {
		case 1:
			adicionarSite()
		case 2:
			removerSite()
		case 3:
			listarSites()
		case 0:
			return
		default:
			color.Red("Opção inválida. Tente novamente.")
		}
	}
}

func adicionarSite() {
	var novoSite string
	fmt.Print("Digite a URL do site a ser adicionado: ")
	fmt.Scan(&novoSite)
	sitesParaMonitorar = append(sitesParaMonitorar, novoSite)
	color.Green("Site adicionado com sucesso: %s\n", novoSite)
}

func removerSite() {
	var siteParaRemover string
	fmt.Print("Digite a URL do site a ser removido: ")
	fmt.Scan(&siteParaRemover)
	for i, site := range sitesParaMonitorar {
		if site == siteParaRemover {
			sitesParaMonitorar = append(sitesParaMonitorar[:i], sitesParaMonitorar[i+1:]...)
			color.Green("Site removido com sucesso: %s\n", siteParaRemover)
			return
		}
	}
	color.Red("Site não encontrado: %s\n", siteParaRemover)
}

func listarSites() {
	color.Cyan("Sites atualmente monitorados:")
	for _, site := range sitesParaMonitorar {
		fmt.Println("-", site)
	}
}

func iniciarMonitoramento(sites []string) {
	fmt.Println("Iniciando monitoramento...")
	for _, site := range sites {
		if err := checarSite(site); err != nil {
			fmt.Println("Erro ao acessar o site:", err)
		}
	}
}

func checarSite(site string) error {
	resp, err := http.Get(site)
	if err != nil {
		logarStatus(site, "Erro ao acessar: "+err.Error())
		return err
	}
	defer resp.Body.Close()
	estadoAtual := resp.StatusCode == http.StatusOK
	if estadoAnterior, existe := statusSites[site]; existe {
		if estadoAtual && !estadoAnterior {
			logarStatus(site, "Página VOLTOU")
		} else if !estadoAtual && estadoAnterior {
			logarStatus(site, "Página CAIU")
		}
	} else {
		if estadoAtual {
			logarStatus(site, "Página UP")
		} else {
			logarStatus(site, "Página DOWN")
		}
	}
	statusSites[site] = estadoAtual
	return nil
}

func logarStatus(site, status string) {
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("%s - %s: %s\n", timestamp, site, status)
	if _, err := arquivoLog.WriteString(logEntry); err != nil {
		fmt.Println("Erro ao escrever no log:", err)
	}
}

func sitesFastShop() []string {
	return []string{
		"https://site.fastshop.com.br/",
		"https://www.fastshop.com.br/",
	}
}

func sitesMCSab() []string {
	return []string{
		"https://www.mcsab.com.br/",
	}
}

func exibirLogs() {
	arquivoLog.Seek(0, 0) // Resetar o ponteiro do arquivo para o início
	scanner := bufio.NewScanner(arquivoLog)
	fmt.Println("Logs de Monitoramento:")
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Erro ao ler o arquivo de log:", err)
	}
}

func limparLogs() {
	arquivoLog.Close()
	err := os.Remove("monitoramento.log")
	if err != nil {
		color.Red("Erro ao remover o arquivo de log:", err)
		return
	}
	color.Green("Logs limpos com sucesso.")

	arquivoLog, err = os.OpenFile("monitoramento.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo de log:", err)
		return
	}
}

func limparTela() {
	var comando string
	if runtime.GOOS == "windows" {
		comando = "cmd.exe"
	} else {
		comando = "clear"
	}
	cmd := exec.Command(comando)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func exibirRodape() {
	fmt.Println("=========================")
}
