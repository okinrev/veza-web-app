# Migration Backend - Guide

## Vue d'ensemble
Migration de l'architecture monolithique vers une architecture clean et scalable.

## Changements principaux

### Structure
- Code organisé dans `internal/` (privé) et `pkg/` (réutilisable)
- Séparation claire admin/api/models
- Couches distinctes: handlers → services → repository

### Nouvelles fonctionnalités
- ✅ Système de permissions basé sur les rôles
- ✅ Réponses API standardisées  
- ✅ Configuration centralisée
- ✅ Architecture modulaire

## Statut de migration

### ✅ Phase 1: Structure et fondations
- [x] Nouvelle structure de dossiers
- [x] Système de réponses API
- [x] Configuration centralisée
- [x] Rôles et permissions

### 🔄 Phase 2: Migration handlers (en cours)
- [x] Handler admin/products adapté
- [ ] Migration autres handlers
- [ ] Tests de compatibilité

### ⏳ Phase 3: Services (à venir)
- [ ] Couche service
- [ ] Couche repository  
- [ ] Tests unitaires

## Test
```bash
./scripts/test-migration.sh
```

## Rollback
Si problème: `git checkout main`
