# ELibrary-NewsService

## Publiczny dostęp
https://jonaszor.github.io/eBiblioteka

## Pobieranie i konfiguracja 

### Pobieranie Go

1. Przejdź na stronę oficjalnej witryny Go: https://golang.org
2. Kliknij na przycisk "Downloads" (Pobierz).
3. Wybierz wersję Go odpowiednią dla swojego systemu operacyjnego.
4. Pobierz instalator i wykonaj instrukcje instalacji.
5. Sprawdzanie instalacji Go.

### Konfiguracja
1. Wprowadź polecenie go version i naciśnij Enter.
Jeśli wszystko zostało poprawnie zainstalowane, powinieneś zobaczyć informacje o wersji Go.
2. Jeśli komenda "go" nie jest rozpoznawana przez środowisko systemowe konieczne będzie ustawienie zmiennych GOROOT, GOBIN oraz GOPATH. Konfiguracja dla środowiska Linux wygląda następująco:
- export GOROOT=/usr/lib/go
- export GOPATH=$HOME/go
- export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

Więcej o prawidłowej konfiguracji GO oraz ustawieniu zmiennych dla platformy Windows znajdziesz tu: https://www.golinuxcloud.com/golang-gopath-vs-goroot/.

3. Uruchom polecenie "go env", jeśli polecenie GO zostanie rozpoznane przez system, a wartość zmiennych GOROOT,GOBIN oraz GOPATCH jest poprawnie ustawiona moesz uruchomić program.

### Program źródłowy
1. Sklonuj repozytorium zawierające kod źródłowy programu lub pobierz go jako archiwum ZIP.
2. Uruchom środowisko, w jakim chcesz odpalić projekt (do programów napisanych w języku Golang zalecany jest edytor Visual Studio Code z rozszerzeniem GO).
3. Wpisz prawidłową konfigurację połączenia w pliku configExample.json. Zmień nazwę pliku na "config.json" lub zmień ściezkę do pliku konfiguracyjnego wewnątrz config.GetConfig().
4. Uruchom program za pomocą polecenia: go run main.go

### Docker
1. Zbuduj obraz Dockera za pomocą polecenia: "docker build -t news-service ." 
2. Uruchom kontener: "docker run -p 8080:8080 news-service"

### Endpointy
Opis endpointów zawierać będzie opis zapytania wraz z jego zawartością oraz wymogami wysłania polecenia, a takze przykład zapytania dla następujących danych:
- Port serwera: 8080,
- Adres serwera: http://localhost:8080, 
- Token JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJncmFudF90eXBlIjoiYWRtaW4iLCJpZCI6IjEifQ.UiFz3OpXIovnhSLKP4h3sMC9ACrWu4_6KEwBPZHI_jo,
- Rola: admin.

#### GET - /api/News
Zapytanie to umoliwia pobranie wszystkich wpisów tablicy News z bazy. Nie wymaga przesłania danych uwierzytelniających w postaci tokenu JWT, tym samym jest dostępny dla wszystkich uytkowników.

Przykładowe polecenie: GET http://localhost:8080/api/News

#### GET - /api/News/{id}
Zapytanie to umoliwia pobranie pojedynczego, wybranego wpisu z tablicy News z bazy. Nie wymaga przesłania danych uwierzytelniających w postaci tokenu JWT, wymaga natomiast podania identyfikatora Id wpisu: jeśli wartość ta występuje w bazie, zostaną pobrane wszystkie dane dla wpisu o podanej wartości, jeśli nie pobranie wdanych nie będzie moliwe..

Przykładowe polecenie: GET http://localhost:8080/api/News/{3}

#### POST - /api/News
Zapytanie to umoliwia utworzenie nowego wpisu News i wysłanie go do bazy. Wymaga podania tokenu JWT zawierającego Id tworzącego wpis oraz rolę, jaką posiada. Do podania tokenu nalezy w sekcji Headers utworzyć pole "Authorization", a w nim umieścić token w postaci "Beaer {token}". W sekcji body naley umieścić zawartość dla pola "Content" odpowiadające treści wpisu.

Przykładowe polecenie: POST http://localhost:8080/api/News

Headers:
- Key: Authorization
- Value: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJncmFudF90eXBlIjoidXNlciIsImlkIjoiMSJ9.Scom-QpoZHVfRdUlLPnVtNCrTm2hhm9mfOWMC_2mZ3w

Body->Raw->JSON:
```json
{
    "Content": "New news example"
}
```

#### PUT - /api/News/{id}
Zapytanie to umoliwia modyfikację wpisu News i uaktualnienie go w bazie. Wymaga podania tokenu JWT zawierającego Id tworzącego wpis oraz rolę, jaką posiada. Do podania tokenu nalezy w sekcji Headers utworzyć pole "Authorization", a w nim umieścić token w postaci "Beaer {token}". W sekcji body naley umieścić zawartość dla pola "Content" odpowiadające treści wpisu. Dodatkowo wymaga podania identyfikatora wpisu, który modyfikujemy.

Przykładowe polecenie: PUT http://localhost:8080/api/News/{3}

Headers:
- Key: Authorization
- Value: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJncmFudF90eXBlIjoidXNlciIsImlkIjoiMSJ9.Scom-QpoZHVfRdUlLPnVtNCrTm2hhm9mfOWMC_2mZ3w

Body->Raw->JSON:
```json
{
    "Content": "Updated news content example"
}
```

#### DELETE - /api/News/{id}
Zapytanie to umoliwia usunięcie wpisu z bazy. Wymaga podania tokenu JWT zawierającego Id tworzącego wpis oraz rolę, jaką posiada. Do podania tokenu nalezy w sekcji Headers utworzyć pole "Authorization", a w nim umieścić token w postaci "Beaer {token}". Wymaga podania identyfikatora wpisu, który usuwamy.

Przykładowe polecenie: DELETE http://localhost:8080/api/News/{3}

Headers:
- Key: Authorization
- Value: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJncmFudF90eXBlIjoidXNlciIsImlkIjoiMSJ9.Scom-QpoZHVfRdUlLPnVtNCrTm2hhm9mfOWMC_2mZ3w