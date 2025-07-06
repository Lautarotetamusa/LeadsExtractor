import matplotlib.pyplot as plt
import random
import string

out_path = "pdfs/static"


def grafico_pagos(c: dict):
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

    # mes, label
    etapas = {
        4: '0%\nArranque de obra',
        int(3 + (c['meses'] + 2) * 0.35): '35%\nEstructura',
        int(3 + (c['meses'] + 2) * 0.55): '55%\nObra gris',
        int(3 + (c['meses'] + 2) * 0.90): '95%\nAcabados',
        3 + c['meses'] + 2: '100%\nFin de obra'
    }

    # Crear la figura y el eje
    fig, ax = plt.subplots(figsize=(12, 3))
    ax.set_ylim(0, max(montos)*2)

    # Graficar las líneas verticales
    for i in range(len(meses)):
        ax.plot([meses[i], meses[i]], [0, montos[i]],
                color='#4c72b0',
                linewidth=2
                )

    # Agregar los textos con los montos sobre los puntos
    for i, monto in enumerate(montos):
        ax.text(meses[i], monto, f"${monto}",
                ha='center',
                va='bottom',
                fontsize=5,
                fontweight='bold',
                color='#333333'
                )

    # Configurar el eje X con etiquetas más legibles
    ax.set_xticks(meses)  # Poner los meses en el eje X

    # Etiquetas para cada mes
    ax.set_xticklabels([str(m) for m in meses], fontsize=10, color='#333333')

    # Eliminar el eje Y completamente
    ax.get_yaxis().set_visible(False)

    # Eliminar título
    ax.set_title('')

    # Eliminar el recuadro alrededor del gráfico
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)
    ax.spines['bottom'].set_visible(False)

    for mes, label in etapas.items():
        ax.text(mes, max(montos)*1.5, label,
                ha='center',
                va='bottom',
                fontsize=9,
                fontweight='bold',
                color='#333333'
                )

        ax.plot([mes, mes], [c['pagos_mensuales']*1.5, max(montos)*1.5],
                color='#333333',
                linestyle='--',
                linewidth=1.5
                )

    # Mejorar el estilo de la gráfica (sin borde y título)
    plt.tight_layout()

    # Guardar el gráfico en un archivo de alta resolución
    path = f"/app/pdfs/static/grafico_pagos.png"
    plt.savefig(path, dpi=300, bbox_inches="tight")

    # Cerrar la figura para liberar recursos
    plt.close()
    return path
