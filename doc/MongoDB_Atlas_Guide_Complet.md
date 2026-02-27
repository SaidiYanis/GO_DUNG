# 📘 MongoDB Atlas -- Guide Complet (Débutant)

Ce guide explique pas à pas comment :

1.  Créer un compte sur MongoDB Atlas\
2.  Créer un cluster gratuit\
3.  Créer une base de données\
4.  Créer un utilisateur pour cette base

------------------------------------------------------------------------

# 1️⃣ Créer un compte sur MongoDB Atlas

1.  Aller sur : https://www.mongodb.com/
2.  Cliquer sur **Try Free** ou **Sign In**
3.  Créer un compte :
    -   Avec Google
    -   Ou avec email + mot de passe
4.  Vérifier votre email si demandé.
5.  Connectez-vous à MongoDB Atlas.

------------------------------------------------------------------------

# 2️⃣ Créer un Cluster Gratuit

Une fois connecté :

1.  Cliquer sur **Build a Database**
2.  Choisir :
    -   ✅ **Shared (Free)** → M0 (gratuit)
3.  Sélectionner :
    -   Cloud Provider : AWS (recommandé)
    -   Region : la plus proche de vous
4.  Cliquer sur **Create Cluster**

⏳ La création peut prendre quelques minutes.

------------------------------------------------------------------------

# 3️⃣ Autoriser votre IP (Obligatoire)

Avant de se connecter à la base :

1.  Aller dans **Network Access**
2.  Cliquer sur **Add IP Address**
3.  Pour développement :
    -   Ajouter `0.0.0.0/0` (autorise toutes les IPs) ⚠️ À éviter en
        production
4.  Cliquer sur **Confirm**

------------------------------------------------------------------------

# 4️⃣ Créer un utilisateur MongoDB

1.  Aller dans **Database Access**
2.  Cliquer sur **Add New Database User**
3.  Choisir :
    -   Authentication Method : Password
4.  Définir :
    -   Username
    -   Password ⚠️ Notez bien le mot de passe !
5.  Database User Privileges :
    -   Sélectionner **Read and Write to Any Database**
6.  Cliquer sur **Add User**

------------------------------------------------------------------------

# 5️⃣ Créer une Base de Données

1.  Aller dans **Database → Browse Collections**
2.  Cliquer sur **Add My Own Data**
3.  Définir :
    -   Database Name (ex: dungeon_game)
    -   Collection Name (ex: players)
4.  Cliquer sur **Create**

🎉 Votre base est créée !

------------------------------------------------------------------------

# 6️⃣ Récupérer l'URL de Connexion

1.  Aller dans **Database**
2.  Cliquer sur **Connect**
3.  Choisir **Drivers**
4.  Sélectionner :
    -   Driver : Go
    -   Version : la plus récente
5.  Copier l'URI de connexion :

```{=html}
<!-- -->
```
    mongodb+srv://<username>:<password>@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority

Remplacer : - `<username>` par votre utilisateur - `<password>` par
votre mot de passe (URL encodé si caractères spéciaux)

------------------------------------------------------------------------

# 7️⃣ Exemple d'URI complète

    mongodb+srv://myuser:mypassword@cluster0.xxxxx.mongodb.net/dungeon_game?retryWrites=true&w=majority&authSource=admin

------------------------------------------------------------------------

# ⚠️ Problèmes Courants

### ❌ Authentication failed

-   Mauvais mot de passe
-   Mot de passe non URL encodé
-   Mauvais utilisateur

### ❌ IP not whitelisted

-   Vérifier Network Access

### ❌ Timeout

-   Vérifier connexion internet
-   Vérifier région cluster

------------------------------------------------------------------------

# 🧠 Bonnes Pratiques

-   Ne jamais commiter l'URI dans Git
-   Utiliser des variables d'environnement
-   Restreindre IP en production
-   Créer des rôles spécifiques par base en production

------------------------------------------------------------------------

# 🎯 Résumé

✔ Compte MongoDB\
✔ Cluster M0 créé\
✔ IP autorisée\
✔ User créé\
✔ Database créée\
✔ URI récupérée

Vous êtes prêt à connecter votre application 🚀
