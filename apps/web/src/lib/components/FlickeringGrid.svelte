<script lang="ts">
  import { onMount } from "svelte";

  type Props = {
    color?: string;
    maxOpacity?: number;
    className?: string;
  };

  let {
    color = "#07a2b6",
    maxOpacity = 0.35,
    className = "",
  }: Props = $props();

  let canvas: HTMLCanvasElement;

  onMount(() => {
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const dpr = window.devicePixelRatio || 1;
    const gap = 26;
    let cols = 0;
    let rows = 0;
    const dots: { x: number; y: number; phase: number; speed: number }[] = [];

    function resize() {
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
      cols = Math.ceil(w / gap);
      rows = Math.ceil(h / gap);
      dots.length = 0;
      for (let r = 0; r <= rows; r++) {
        for (let c = 0; c <= cols; c++) {
          dots.push({
            x: c * gap,
            y: r * gap,
            phase: Math.random() * Math.PI * 2,
            speed: 0.4 + Math.random() * 0.9,
          });
        }
      }
    }

    let raf = 0;
    let t = 0;

    function draw() {
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      ctx!.clearRect(0, 0, w, h);
      ctx!.fillStyle = color;
      for (const d of dots) {
        const o = reduceMotion
          ? maxOpacity * 0.4
          : (Math.sin(t * d.speed + d.phase) * 0.5 + 0.5) * maxOpacity;
        ctx!.globalAlpha = o;
        ctx!.beginPath();
        ctx!.arc(d.x, d.y, 1.3, 0, Math.PI * 2);
        ctx!.fill();
      }
      ctx!.globalAlpha = 1;
      if (!reduceMotion) {
        t += 0.03;
        raf = requestAnimationFrame(draw);
      }
    }

    resize();
    draw();
    window.addEventListener("resize", resize);
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("resize", resize);
    };
  });
</script>

<canvas bind:this={canvas} class={className} aria-hidden="true"></canvas>
