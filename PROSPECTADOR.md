# Propespectador de leads

Sistema de generación de leads para **Rebora** a partir del portal [inmuebles24.com](https://www.inmuebles24.com). Extrae propiedades de alto valor, identifica a las inmobiliarias que las publican y les envía una secuencia automatizada de mensajes por WhatsApp con el objetivo de convertirlas en clientes de Rebora.

---

## Tabla de contenidos

1. [Propósito y contexto](#1-propósito-y-contexto)

---

## 1. Propósito y contexto

Rebora es una inmobiliaria especializada en propiedades residenciales de alto valor en Guadalajara, México. LeadHunter automatiza la prospección de nuevos clientes (inmobiliarias que ya tienen propiedades publicadas) mediante el siguiente ciclo:

1. **Scraping** de propiedades en inmuebles24 con filtro de precio mínimo ($250,000,000 MXN).
2. **Identificación** de la inmobiliaria publicante (lead) de cada propiedad.
3. **Secuencia de 7 mensajes por WhatsApp**, uno por semana por cada propiedad, enviados de mayor a menor precio.
4. **Terminación anticipada** si la inmobiliaria responde afirmativamente (interés real de co-listado).
