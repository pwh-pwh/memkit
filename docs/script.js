const cards = document.querySelectorAll(".card, .mode");

const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("reveal");
        observer.unobserve(entry.target);
      }
    });
  },
  { threshold: 0.2 }
);

cards.forEach((card, idx) => {
  card.style.setProperty("--delay", `${idx * 80}ms`);
  observer.observe(card);
});

const code = document.querySelectorAll(".code-window code, .terminal-body code");
code.forEach((block) => {
  const content = block.textContent;
  block.textContent = "";
  let i = 0;
  const step = () => {
    block.textContent += content[i];
    i += 1;
    if (i < content.length) {
      requestAnimationFrame(step);
    }
  };
  step();
});
