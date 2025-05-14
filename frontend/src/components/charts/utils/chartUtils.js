export function generateColors(count) {
  const colors = [
    '#4e79a7', '#f28e2b', '#e15759', '#76b7b2',
    '#59a14f', '#edc949', '#af7aa1', '#ff9da7',
  ];
  return Array.from({ length: count }, (_, i) => colors[i % colors.length]);
}

export function defaultOptions(title) {
  return {
    responsive: true,
    plugins: {
      title: {
        display: true,
        text: title,
      },
      legend: {
        position: 'top',
      },
    },
  };
}
