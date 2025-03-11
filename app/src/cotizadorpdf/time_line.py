import matplotlib.pyplot as plt
import random
import string

def generar_nombre(longitud=8):
    caracteres = string.ascii_letters + string.digits  # Letras y números
    return ''.join(random.choice(caracteres) for _ in range(longitud))


def grafico_etapas(c):
    meses = list(range(0, c['meses'] + 4))  # Meses desde 0 hasta el máximo
    
    fig, ax = plt.subplots(figsize=(12, 4))
    
    # Configurar el eje X con etiquetas
    ax.set_xticks(meses)
    ax.set_xticklabels([str(m) for m in meses], fontsize=10, color='#333333')
    ax.set_xlim(-0.5, c['meses'] + 4.5)
    
    # Eliminar el eje Y
    ax.get_yaxis().set_visible(False)
    
    # Eliminar bordes
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)
    ax.spines['bottom'].set_visible(False)
    
    # Etapas de construcción
    etapas = {
        3: '0%\nArranque de obra',
        int(3 + c['meses'] * 0.35): '35%\nEstructura',
        int(3 + c['meses'] * 0.55): '55%\nObra gris',
        int(3 + c['meses'] * 0.90): '95%\nAcabados',
        3 + c['meses']: '100%\nFin de obra'
    }
    
    for mes, label in etapas.items():
        ax.text(mes, 0.01, label, ha='center', va='bottom', fontsize=9, fontweight='bold', color='#333333')
        ax.plot([mes, mes], [0, 0.01], color='#333333', linestyle='--', linewidth=1.5)
    
    plt.tight_layout()
    
    nombre = generar_nombre() + ".png"
    plt.savefig(nombre, dpi=300, bbox_inches="tight", transparent=True)
    plt.close()
    
    return nombre

def grafico_pagos(c):
    # Datos de ejemplo: Meses (1 al 21) y los montos a pagar
    meses = list(range(0, c['meses']+4))  # Meses del 0 al 21
    montos = []
    montos.append(c['anticipo'])
    montos.append(0)
    montos.append(0)
    montos.append(c['valor_previo'])
    for _ in range(1, c['meses']+1):
        montos.append(c['pagos_mensuales'])

    # Crear la figura y el eje
    fig, ax = plt.subplots(figsize=(12, 4))

    # Graficar las líneas verticales
    for i in range(len(meses)):
        ax.plot([meses[i], meses[i]], [0, montos[i]], color='#4c72b0', linewidth=2)  # Línea vertical

    # Agregar los textos con los montos sobre los puntos
    for i, monto in enumerate(montos):
        ax.text(meses[i], monto + 700, f"${monto}", ha='center', va='bottom', fontsize=7, fontweight='bold', color='#333333')

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

    # Añadir el texto para las fases de la construcción
    ax.text(1.5, max(montos) * 0.85, 'Preconstrucción', ha='center', va='bottom', fontsize=10, fontweight='bold', color='#333333')
    ax.text(3+c['meses']/2, max(montos) * 0.85, 'Construcción', ha='center', va='bottom', fontsize=10, fontweight='bold', color='#333333')

    # Dibujar una línea divisoria entre las dos fases
    ax.plot([3.5, 3.5], [0, max(montos) * 0.85], color='#333333', linestyle='--', linewidth=2)

    # Mejorar el estilo de la gráfica (sin borde y título)
    plt.tight_layout()

    # Guardar el gráfico en un archivo de alta resolución
    nombre = generar_nombre()
    nombre = nombre +".png"
    plt.savefig( nombre, dpi=300, bbox_inches="tight", transparent=True)


    # Cerrar la figura para liberar recursos
    plt.close()
    return nombre