# Déploiement de gatus-custom (infologique)

Notes pour reconstruire et redéployer **ce fork** de Gatus tel qu'il tourne en production.
À lire par un humain ou un agent avant de toucher au conteneur.

## Où ça tourne

| Élément | Valeur |
|---|---|
| Serveur | `flatcar01` — 192.168.1.61 (Flatcar Container Linux, VM Hyper-V) |
| Dossier du projet | `/mnt/data/gatus` (clone de ce dépôt, appartient à root → utiliser `sudo`) |
| Partition | `/mnt/data` = `/dev/sda10` (label `DOCKERDATA`), montée par `mnt-data.mount` |
| Conteneur | `gatus`, image locale `gatus-custom:latest`, `restart: unless-stopped` |
| Port | **8081** sur l'hôte → 8080 dans le conteneur (8080 est pris par AdGuard Home) |
| Accès public | https://status.infologique.net → nginx sur 192.168.1.38 (`/etc/nginx/conf.d/status.infologique.net.conf`) → `192.168.1.61:8081` |
| Historique | `./data/gatus.db` (SQLite, monté dans `/data`) — **ne pas supprimer** |
| Config | `./config.yaml` monté dans `/config/config.yaml` (rechargé à chaud par Gatus) |

Historique : jusqu'au 2026-10-12, tournait sur l'ancien serveur 192.168.1.148 dans `/root/gatus-custom/gatus` (port 8080).

## Modifier seulement la config

Pas besoin de rebuild : Gatus surveille `config.yaml` et se recharge tout seul.

```bash
ssh anisite@192.168.1.61
cd /mnt/data/gatus
sudo vim config.yaml
sudo docker logs -f gatus      # vérifier qu'il n'y a pas d'erreur de validation (une config invalide fait planter Gatus)
```

## Reconstruire l'image après un changement de code

```bash
ssh anisite@192.168.1.61
cd /mnt/data/gatus
sudo git pull                          # ou modifier le code directement ici
sudo docker tag gatus-custom:latest gatus-custom:previous   # garder l'image actuelle pour un retour arrière
sudo docker compose up -d --build      # compile (golang:alpine, multi-stage) puis remplace le conteneur
sudo docker logs --tail 20 gatus
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8081/   # doit répondre 200
```

- La compilation se fait entièrement dans Docker (`Dockerfile` multi-stage) : aucun Go à installer sur l'hôte.
- Si le code touche `web/app/`, lancer `make frontend-build` **avant** le build : `web/static/` est embarqué dans le binaire (voir `AGENTS.md`).
- `golang:alpine` n'est pas épinglé. Si un build casse à cause d'une nouvelle version de Go, fixer la version dans le `Dockerfile` (ex. `golang:1.24-alpine`).
- Gatus est coupé seulement le temps de remplacer le conteneur (quelques secondes) ; l'historique est conservé.

### Retour arrière

```bash
sudo docker tag gatus-custom:previous gatus-custom:latest
sudo docker compose up -d --no-build
```

## Déplacer Gatus sur un autre serveur

1. Arrêter l'ancien conteneur **avant** de copier `data/` (base SQLite cohérente, et pas d'alertes ntfy en double).
2. Copier tout le dossier (sources + `config.yaml` + `data/`), par ex. `tar -cf - gatus | ssh hôte 'sudo tar -xpf - -C /mnt/data'`.
3. Soit rebuild (`docker compose up -d --build`), soit transférer l'image telle quelle :
   `docker save gatus-custom:latest | gzip | ssh hôte 'gunzip | sudo docker load'` puis `docker compose up -d --no-build`.
4. Mettre à jour le `proxy_pass` de `status.infologique.net` sur le nginx (192.168.1.38), `sudo nginx -t`, puis `sudo systemctl reload nginx`.

## Attention : dépôt public

Ce dépôt est **public**. `config.yaml` contient des URL de sondes qui ne devraient idéalement pas l'être
(`codeConfirmation` de Mon dossier, clé dans l'URL RSS de SAG, sujets ntfy). Ne pas y ajouter de nouveaux secrets.
