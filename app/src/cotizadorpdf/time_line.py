import matplotlib.pyplot as plt
import random
import string

def generar_nombre(longitud=8):
    caracteres = string.ascii_letters + string.digits  # Letras y números
    return ''.join(random.choice(caracteres) for _ in range(longitud))


def grafico_etapas(c):
    meses = list(range(0, c['meses'] + 4))  # Meses desde 0 hasta el máximo
    
    fig, ax = plt.subplots(figsize=(12, 2))
    
    # Configurar el eje X con etiquetas
    ax.set_xticks(meses)
    ax.set_xticklabels([str(m) for m in meses], fontsize=10, color='#333333')
    ax.set_xlim(-1.2, c['meses'] + 4.5)
    
    # Eliminar el eje Y
    ax.set_ylim(0, 1) 
    
    # Eliminar bordes
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)
    ax.spines['bottom'].set_visible(False)
    
    # Etapas de construcción
    etapas = {
        4: '0%\nArranque de obra',
        int(3 + c['meses'] * 0.35): '35%\nEstructura',
        int(3 + c['meses'] * 0.55): '55%\nObra gris',
        int(3 + c['meses'] * 0.90): '95%\nAcabados',
        3 + c['meses']: '100%\nFin de obra'
    }
    
    for mes, label in etapas.items():
        ax.text(mes, 1, label, ha='center', va='bottom', fontsize=9, fontweight='bold', color='#333333')
        ax.plot([mes, mes], [0.5, 1], color='#333333', linestyle='--', linewidth=1.5)
    
    ax.yaxis.set_ticks([])  # Quita los ticks del eje Y
    ax.yaxis.set_ticklabels([])  # Quita las etiquetas del eje Y

    ax.xaxis.set_tick_params(pad=1)
    ax.xaxis.set_ticks_position('bottom')  # Coloca los ticks en la parte inferior
    ax.tick_params(axis='x', which='both', direction='inout', length=6, width=2)
    
    plt.tight_layout()
    
    nombre = generar_nombre() + ".png"
    plt.savefig(nombre, dpi=300, bbox_inches="tight", transparent=True)
    plt.close()
    
    return nombre

def grafico_pagos(c):
    # Datos de ejemplo: Meses (1 al 21) y los montos a pagar
    meses = list(range(0, c['meses']+6))  # Meses del 0 al 21
    montos = []
    montos.append(c['anticipo'])
    montos.append(0)
    montos.append(0)
    montos.append(c['valor_previo'])
    for _ in range(1, c['meses']+1):
        montos.append(c['pagos_mensuales'])
    montos.append(0)
    montos.append(0)
    # Crear la figura y el eje
    fig, ax = plt.subplots(figsize=(12, 3))

    ax.set_ylim(0, max(montos)*2) 

    # Graficar las líneas verticales
    for i in range(len(meses)):
        ax.plot([meses[i], meses[i]], [0, montos[i]], color='#4c72b0', linewidth=2)  # Línea vertical

    # Agregar los textos con los montos sobre los puntos
    for i, monto in enumerate(montos):
        ax.text(meses[i], monto, f"${monto}", ha='center', va='bottom', fontsize=5, fontweight='bold', color='#333333')

    # Configurar el eje X con etiquetas más legibles
    ax.set_xticks(meses)  # Poner los meses en el eje X
    ax.set_xticklabels([str(m) for m in meses], fontsize=10, color='#333333')  # Etiquetas para cada mes

    # Eliminar el eje Y completamente
    ax.get_yaxis().set_visible(False)

    # Eliminar título
    ax.set_title('')  # Sin título

    # Eliminar el recuadro alrededor del gráfico
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)
    ax.spines['bottom'].set_visible(False)

    etapas = {
            4: '0%\nArranque de obra',
            int(3 + (c['meses'] + 2) * 0.35): '35%\nEstructura',
            int(3 + (c['meses'] + 2) * 0.55): '55%\nObra gris',
            int(3 + (c['meses'] + 2) * 0.90): '95%\nAcabados',
            3 + c['meses'] + 2: '100%\nFin de obra'
    }
    for mes, label in etapas.items():
        ax.text(mes, max(montos)*1.5, label, ha='center', va='bottom', fontsize=9, fontweight='bold', color='#333333')
        ax.plot([mes, mes], [c['pagos_mensuales']*1.5, max(montos)*1.5], color='#333333', linestyle='--', linewidth=1.5)
    # Mejorar el estilo de la gráfica (sin borde y título)
    plt.tight_layout()

    # Guardar el gráfico en un archivo de alta resolución
    nombre = generar_nombre()
    nombre = nombre +".png"
    plt.savefig( nombre, dpi=300, bbox_inches="tight", transparent=True)


    # Cerrar la figura para liberar recursos
    plt.close()
    return nombre