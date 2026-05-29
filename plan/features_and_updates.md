# Features and Updates Summary

This document provides an overview of the key features, updates, and improvements implemented in the repository, compiled from recent commit logs.

## 1. Expanded AI Reports & Multi-LLM Provider Integration
* **LLM Provider Expansion**: Wired in backend and frontend support for multiple AI/LLM providers (including OpenAI, OpenRouter, Ollama, Anthropic, and Gemini) along with a provider-neutral configuration page.
* **Smart Bird Links & Media Fallbacks**:
  * Fixed eBird link generation by resolving missing species codes from the BirdNET offline taxonomy (rather than using hyphenated scientific names).
  * Grouped/stacked external bird links in AI reports.
  * Added fallback dashboard-styled SVG initials for bird species that are missing thumbnail images.
* **Report Access Control**:
  * Configured access rules so that public/guest users can view cached AI reports without logging in, while preventing them from triggering expensive report refreshes or cache bypasses.
  * Added a configuration setting for AI report days.
  * Added UTM tracking parameters to external links inside AI reports.

## 2. Visitor Analytics & Admin Dashboard
* **Visitor Tracking System**: Implemented tracking of client-side SPA page views, public page visits, and entry referrers, writing them to a structured log file.
* **Admin Dashboard**: Created an admin dashboard to visualize visitor logs and analyze traffic.
* **AI Report Statistics**: Added view counters specifically to track and show how many times the AI reports are viewed.

## 3. Core Backend Improvements & Configuration Hardening
* **Config Secrets Encryption**: Added automatic encryption and decryption of API keys, passwords, and sensitive settings in `config.yaml` using AES-GCM (with environment variable or key-file resolver fallbacks).
* **Dependency Wiring**: Added missing/untracked packages for critical subsystems: `tflite`, `ffmpeg`, and `rtsp`.
* **Frontend Compatibility**: Resolved TypeScript type-checking and compilation issues on the frontend settings page.
* **Backend Cleanup**: Cleaned up routes, authentication logic, and settings syncing in backend controllers.

## 4. Documentation & Guides
* **Thumbnail Plan**: Documented a plan for thumbnail link integration.
* **Developer Handbooks**: Created an AI Assistant handoff guide, AI Analyzer changelog, and updated the main README with documentation links.
