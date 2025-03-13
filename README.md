# Projeto Auction

Este projeto é um sistema de leilões desenvolvido em Go, utilizando MongoDB como banco de dados.

## Pré-requisitos

- Docker e Docker Compose instalados.
- Go instalado (para rodar testes e desenvolver localmente sem Docker, se desejado).

## Rodando em Ambiente de Desenvolvimento

1. Navegue até a raiz do projeto:
   ```bash
   cd /Users/andrefelizardo/Documents/devfelizardo/Go/posgoexpert_labs-auction
   ```
2. Utilize o Docker Compose para construir e iniciar os serviços:

   ```bash
   docker compose up --build
   ```

   - O serviço `app` será iniciado na porta `8080`.
   - O MongoDB será iniciado na porta `27017`.

3. Para visualizar os logs do aplicativo:
   ```bash
   docker compose logs -f app
   ```

## Configuração de Variáveis de Ambiente

- As variáveis de ambiente estão definidas no arquivo `cmd/auction/.env`.
- Exemplos:
  - `AUCTION_DURATION` – define a duração dos leilões.
  - Outras configurações necessárias para a aplicação.

## Rodando os Testes

Para executar os testes de forma local, execute na raiz do projeto:

```bash
go test ./...
```

## Parando os Serviços

Para parar e remover os containers, execute:

```bash
docker compose down
```

## Testando os Endpoints de Forma Assertiva

O arquivo main.go define os seguintes endpoints:

- GET /auction – lista todos os leilões;
- GET /auction/:auctionId – retorna os detalhes de um leilão específico;
- POST /auction – cria um novo leilão;
- GET /auction/winner/:auctionId – retorna o lance vencedor de um leilão;
- POST /bid – cria um novo lance para um leilão;
- GET /bid/:auctionId – lista os lances de um leilão;
- GET /user/:userId – retorna os dados de um usuário.

### Como Testar

1. **Testes Manuais com Ferramentas:**  
   Utilize Postman, Insomnia ou cURL para enviar requisições e validar os seguintes pontos:

   - O código de status HTTP (por exemplo, 200, 201 ou 400).
   - O formato e os dados do JSON retornado.
   - O fluxo de criação, consulta e atualização do status do leilão, verificando que após o intervalo definido (por exemplo, via AUCTION_DURATION) o leilão é automaticamente encerrado.

   Exemplo com cURL:

   ```bash
   # Criar leilão
   curl -X POST http://localhost:8080/auction \
     -H "Content-Type: application/json" \
     -d '{"product_name": "Smartphone", "category": "Eletrônicos", "description": "Modelo topo", "condition": 1}'

   # Consultar leilão (substitua {auctionId} pelo ID retornado)
   curl http://localhost:8080/auction/{auctionId}
   ```
