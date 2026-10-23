# Mock-EVCC für lokale Demos

`cmd/evccmock` bildet die EVCC-Endpunkte nach, die Ladegast nutzt
(siehe DESIGN.md Abschnitt 3), und liefert Daten aus `evcc-mock.yaml`.
Nur für Entwicklung und Ende-zu-Ende-Demos.

```bash
task evcc:mock            # startet auf :7070
task evcc:mock:generate   # erzeugt evcc-mock.yaml neu (deterministisch, Seed 1)
```

Ladegast anbinden: `LADEGAST_EVCC_BASE_URL=http://localhost:7070`,
`LADEGAST_EVCC_API_KEY=evcc_mock_key`. Freunde mit den Fahrzeugnamen
`vehicle_1` (Anna), `vehicle_2` (Ben), `vehicle_3` (Clara) anlegen.

## Endpunkte

| Methode | Pfad | Verhalten |
|---|---|---|
| GET | `/api/sessions?year=&month=` | Sessions aller Fahrzeuge, aufsteigend nach `created` |
| POST | `/api/vehicles/{name}/mode/{mode}` | setzt den Modus im Speicher (`off`, `now`, `minpv`, `pv`, `smart`) |
| GET | `/api/vehicles/{name}/mode` | nur Mock: aktuellen Modus abfragen |

Authentifizierung: `Authorization: Bearer <apiKey>` (leerer `apiKey` = aus).

## Dateiformat (`evcc-mock.yaml`)

- `listen`, `apiKey`, `timezone` (für den year/month-Filter).
- `referenceTime` + `rebaseToNow: true`: beim Start werden alle Zeitstempel um
  ganze Wochen nach vorn geschoben, sodass die Daten kurz vor "jetzt" enden
  (Wochentage bleiben erhalten). Mit `false` werden die Zeitstempel
  unverändert ausgeliefert.
- `users[]` mit genau einem `vehicle` (`name` = EVCC-Config-Name, `title`,
  `capacityKWh`, `loadpoint`, `mode`) und `sessions[]` (`created`, `finished`,
  `chargedEnergy` in kWh, `solarPercentage`, `price`, `pricePerKWh`).
- Unbekannte Schlüssel und ungültige Werte brechen den Start ab.

## Beispieldaten

Drei Nutzer mit je einem Fahrzeug über 12 Monate, überwiegend Solarladen
(Sommer ~95 %, Winter ~60 %, Abendladungen mit wenig Sonne). Jeder hat
mehrwöchige Pausen (Urlaub, Auto steht) und viele einzelne Tage ohne Laden.
