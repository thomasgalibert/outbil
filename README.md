# OutBil - Gestionnaire de Devis Professionnel

OutBil est une CLI élégante et intuitive pour créer et gérer des devis professionnels avec une base de données SQLite locale.

![Demo OutBil](demo-outbil.gif)

## Installation

### Prérequis

- Go 1.19 ou supérieur
- Git (pour cloner le repository)

### Compilation depuis les sources

```bash
# Cloner le repository
git clone <repo>
cd outbil

# Compiler pour votre système
go build -o outbil .

# Ou utiliser make si disponible
make build
```

### Création des exécutables pour différentes plateformes

```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o outbil.exe .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o outbil-macos-intel .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o outbil-macos-arm64 .

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o outbil-linux .
```

### Installation globale

```bash
# Installer l'exécutable dans votre PATH
go install .

# Ou copier manuellement
sudo cp outbil /usr/local/bin/
```

## Configuration initiale

### 1. Informations de votre entreprise

La première étape est de configurer les informations de votre société :

```bash
outbil company setup
```

Vous devrez fournir :
- Nom de l'entreprise
- Adresse complète
- Téléphone
- Email
- Numéro SIRET
- Numéro de TVA intracommunautaire
- Conditions de paiement par défaut

### 2. Logo de l'entreprise (optionnel)

Pour ajouter votre logo sur les devis PDF, placez simplement votre fichier logo à la racine du projet :
- `logo.jpg` ou `logo.jpeg` ou `logo.png`
- Taille recommandée : 200x100 pixels
- Le logo sera automatiquement détecté et ajouté aux PDFs

### 3. Conditions Générales de Vente (optionnel)

Pour joindre automatiquement vos CGV aux devis :
1. Créez un fichier PDF nommé `cgv.pdf`
2. Placez-le à la racine du projet
3. Les CGV seront automatiquement fusionnées à la fin de chaque devis généré

## Utilisation

### Aide et commandes disponibles

```bash
# Aide générale
outbil --help
outbil -h

# Aide sur une commande spécifique
outbil quote --help
outbil client --help
outbil company --help
```

### Gestion des clients

```bash
# Lister tous les clients
outbil client list

# Ajouter un nouveau client (interactif)
outbil client add

# Afficher les détails d'un client
outbil client show <ID>

# Modifier un client
outbil client edit <ID>

# Supprimer un client
outbil client delete <ID>
```

### Gestion des devis

```bash
# Lister tous les devis
outbil quote list

# Créer un nouveau devis (interactif)
outbil quote create

# Afficher les détails d'un devis
outbil quote show <ID>

# Modifier un devis existant
outbil quote edit <ID>

# Dupliquer un devis existant
outbil quote duplicate <ID>

# Modifier le statut d'un devis
outbil quote status <ID>

# Exporter un devis en PDF
outbil quote pdf <ID>

# Supprimer un devis
outbil quote delete <ID>
```

### Gestion de l'entreprise

```bash
# Afficher les informations actuelles
outbil company show

# Modifier les informations
outbil company setup
```

### Gestion des bases de données

```bash
# Lister toutes les bases disponibles
outbil db list

# Voir la base active
outbil db current

# Créer une nouvelle base (ex: demo, test, client1)
outbil db create demo

# Changer de base active
outbil db switch demo

# Supprimer une base (avec confirmation)
outbil db delete test
```

## Fonctionnalités

✨ **Gestion complète des clients** - Créez et gérez votre base de clients facilement

📝 **Création interactive de devis** - Interface intuitive pour ajouter des lignes de devis

📊 **Calcul automatique** - TVA, sous-totaux et totaux calculés automatiquement

📄 **Export PDF professionnel** - Générez des devis avec logo et CGV automatiquement

🔄 **Duplication de devis** - Créez rapidement un nouveau devis basé sur un existant

🏷️ **Identifiants uniques** - Chaînes aléatoires de 8 caractères majuscules (ex: KXPQWMZN)

📈 **Suivi des statuts** - Brouillon, Envoyé, Accepté, Refusé, Expiré

💾 **Base SQLite locale** - Données stockées dans `~/.outbil/outbil.db`

📁 **Organisation des PDFs** - Les devis sont sauvegardés dans le dossier `quotes/`

🗄️ **Bases de données multiples** - Gérez plusieurs bases (production, demo, test)

## Structure des données

### Clients
- Nom et prénom
- Entreprise (optionnel)
- Email
- Téléphone
- Adresse complète
- Numéro de TVA (optionnel)

### Devis
- Numéro unique (8 lettres majuscules)
- Client associé
- Date de création
- Date de validité (1 mois par défaut)
- Statut
- Notes (optionnel)
- Conditions de paiement
- Lignes de produits/services

### Lignes de devis
- Description
- Quantité
- Prix unitaire HT
- Taux de TVA (20% par défaut)
- Montant total HT calculé

## Exemple de workflow complet

1. **Configuration initiale** :
   ```bash
   outbil company setup
   ```

2. **Ajout du logo** (optionnel) :
   ```bash
   cp mon-logo.png logo.png
   ```

3. **Ajout des CGV** (optionnel) :
   ```bash
   cp mes-cgv.pdf cgv.pdf
   ```

4. **Création d'un client** :
   ```bash
   outbil client add
   ```

5. **Création d'un devis** :
   ```bash
   outbil quote create
   ```

6. **Export en PDF** :
   ```bash
   outbil quote pdf 1
   # Le PDF sera créé dans quotes/2024_01_devis_ABCDEFGH.pdf
   ```

7. **Envoi et suivi** :
   ```bash
   # Marquer comme envoyé
   outbil quote status 1
   # Choisir "sent"
   ```

## Stockage des données

- **Bases de données** : `~/.outbil/*.db` (outbil.db par défaut)
- **Base active** : `~/.outbil/config`
- **PDFs générés** : `./quotes/`
- **Logo** : `./logo.{jpg,jpeg,png}`
- **CGV** : `./cgv.pdf`

## Notes importantes

- Les devis sont valides 1 mois par défaut (modifiable)
- La TVA par défaut est de 20% (modifiable par ligne)
- Les PDFs sont protégés contre la modification (impression autorisée)
- Les montants sont arrondis à 2 décimales
- Format des dates : JJ/MM/AAAA
- Format des numéros de devis : AAAA-MM-XXXXXXXX

## Dépannage

### La base de données n'est pas trouvée
```bash
# Vérifier l'emplacement
ls ~/.outbil/

# Recréer si nécessaire
rm -rf ~/.outbil/
outbil company setup
```

### Erreur de compilation
```bash
# Installer les dépendances
go mod download

# Nettoyer et recompiler
go clean
go build .
```

### Le logo n'apparaît pas sur les PDFs
- Vérifiez que le fichier est nommé exactement `logo.jpg`, `logo.jpeg` ou `logo.png`
- Assurez-vous qu'il est dans le répertoire d'exécution d'OutBil
- Format recommandé : PNG avec transparence

## Licence

MIT License - voir le fichier [LICENSE](LICENSE) pour plus de détails.