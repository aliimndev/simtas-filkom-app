import { canAnimateViewTransitionPseudoElement } from "../animated-theme-toggler"

/**
 * Covers the cross-browser capability detection that decides whether the theme
 * reveal runs through the JS animations path (Chromium/Brave, Firefox >= 140)
 * or the CSS keyframe fallback (Firefox < 140, engines lacking pseudoElement
 * support for ::view-transition-*). jsdom has no KeyframeEffect, so we stub the
 * global and exercise each branch.
 */
describe("canAnimateViewTransitionPseudoElement", () => {
  afterEach(() => {
    delete (globalThis as { KeyframeEffect?: unknown }).KeyframeEffect
  })

  it("false when KeyframeEffect is unavailable", () => {
    delete (globalThis as { KeyframeEffect?: unknown }).KeyframeEffect
    expect(canAnimateViewTransitionPseudoElement()).toBe(false)
  })

  it("false when the constructor rejects ::view-transition-* pseudoElement (Firefox < 140)", () => {
    ;(globalThis as { KeyframeEffect?: unknown }).KeyframeEffect = class {
      constructor() {
        throw new DOMException("unsupported pseudo-element", "SyntaxError")
      }
    }
    expect(canAnimateViewTransitionPseudoElement()).toBe(false)
  })

  it("false when the constructed effect does not keep the requested pseudoElement", () => {
    ;(globalThis as { KeyframeEffect?: unknown }).KeyframeEffect = class {
      pseudoElement: string | null = null
    }
    expect(canAnimateViewTransitionPseudoElement()).toBe(false)
  })

  it("true when it accepts the view-transition pseudoElement (Chromium/Brave, Firefox >= 140)", () => {
    ;(globalThis as { KeyframeEffect?: unknown }).KeyframeEffect = class {
      pseudoElement: string | null
      constructor(
        _target: Element | null,
        _keyframes: Keyframe[],
        options: KeyframeEffectOptions
      ) {
        this.pseudoElement = options.pseudoElement ?? null
      }
    }
    expect(canAnimateViewTransitionPseudoElement()).toBe(true)
  })
})