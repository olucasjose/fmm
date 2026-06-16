# FMM (Fastest Mint Mirrors)

🌍 *Leia em outros idiomas: [English](README.md) | [Português](README.pt-BR.md)*

FMM é uma ferramenta rápida e interativa para selecionar mirrors (espelhos) do Linux Mint. Ela testa os mirrors disponíveis por ordem de proximidade geográfica e ajuda você a atualizar a configuração de mirrors do seu sistema para otimizar as velocidades de download.

## Escopo e Filosofia

**O FMM foca exclusivamente no teste e na atualização de mirrors.**

Outros recursos encontrados na ferramenta original `mintsources` — como o gerenciamento de chaves GPG, gerenciamento de PPAs e outras configurações de repositório — **não** fazem parte do escopo deste projeto. Esses recursos não foram implementados e não há planos para implementá-los no futuro. O FMM foi construído para fazer apenas uma coisa e fazê-la extremamente bem: encontrar e configurar os mirrors mais rápidos.

## Origem e Créditos

O FMM usa o projeto original [mintsources](https://github.com/linuxmint/mintsources) como sua base direta. Este projeto é um port em Go e um trabalho derivado da lógica de testes de mirrors do `mintsources`, que é desenvolvido e mantido pela equipe do Linux Mint. O FMM baseia-se nessa estrutura robusta para fornecer uma experiência de linha de comando (CLI) mais rápida.

O FMM é licenciado sob a **GNU General Public License v3.0 (GPLv3)**. O software original `mintsources` é distribuído sob a licença GNU GPL. Consulte o arquivo [LICENSE](LICENSE) para obter mais detalhes.

## Instalação

Você pode instalar o FMM diretamente executando o script de instalação fornecido. O script compila o binário em Go, move-o para o caminho do seu sistema e configura o autocompletar no bash.

```bash
# Clone o repositório
git clone https://github.com/olucasjose/fmm.git
cd fmm

# Execute o script de instalação (requer privilégios sudo)
./install.sh
```

## Como Usar

O FMM exige privilégios `sudo` para ações que modificam as configurações do sistema (como `run`), mas os comandos informativos podem ser executados por um usuário comum.

### Comandos Principais & Flags

- **Execute o fluxo principal de teste e atualização de mirrors:**
  ```bash
  sudo fmm run [flags]
  ```
  **Principais Flags para `run`:**
  - `-a, --apply`: Aplica os mirrors mais rápidos automaticamente no sistema.
  - `-u, --update-cache`: Atualiza o cache do APT (`apt update`) logo após aplicar os mirrors.
  - `-c, --countries <lista>`: Filtra mirrors por um código de país específico (ex: `BR,US`).
  - `-l, --limit <número>`: Limita a quantidade de mirrors que serão testados.
  - `-v, --viable <número>`: Para os testes após encontrar uma quantidade específica de mirrors viáveis.
  - `-t, --target-speed <velocidade>`: Para de testar assim que algum mirror alcançar esta velocidade alvo (ex: `10mbps`).
  - `-e, --show-errors`: Mostra detalhes dos erros de conexão durante os testes.
  - `-q, --quiet`: Oculta toda a saída visual (útil para automação via cronjob).

- **Liste os mirrors disponíveis:**
  ```bash
  fmm list [flags]
  ```
  **Principais Flags para `list`:**
  - `-c, --countries <lista>`: Filtra a lista exibindo apenas os países especificados.
  - `-r, --regions <lista>`: Filtra a lista exibindo apenas as regiões/continentes especificados.

- **Exiba a ajuda e os comandos disponíveis:**
  ```bash
  fmm help
  ```
