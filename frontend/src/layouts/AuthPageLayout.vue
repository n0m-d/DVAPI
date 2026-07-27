<script setup lang="ts">
import ModeToggle from '@/components/dashboard/ModeToggle.vue'
import ThemeCustomizer from '@/components/dashboard/ThemeCustomizer.vue'

defineProps<{
  title: string
  description: string
  showTerms?: boolean
}>()
</script>

<template>
  <div class="auth-page relative min-h-screen overflow-hidden">
    <div class="auth-orb auth-orb-1" />
    <div class="auth-orb auth-orb-2" />
    <div class="auth-orb auth-orb-3" />
    <div class="auth-scanline" />

    <div class="absolute right-4 top-4 z-30 flex gap-2 md:right-8 md:top-8">
      <div class="auth-controls flex items-center gap-1">
        <ThemeCustomizer />
        <ModeToggle />
      </div>
    </div>

    <div class="container relative z-10 flex min-h-screen flex-col items-center justify-center px-4 py-16 lg:grid lg:max-w-none lg:grid-cols-2 lg:gap-12 lg:px-8">
      <div class="relative hidden h-full flex-col justify-between py-8 lg:flex">
        <div class="auth-grid auth-grid-animated absolute inset-0" />

        <div class="relative z-10 flex items-center gap-3">
          <div class="flex size-10 items-center justify-center rounded-lg border border-border bg-muted/50 backdrop-blur-sm">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="auth-accent size-5"
          >
            <path d="M3 21h18" />
            <path d="M5 21V9l7-5 7 5v12" />
            <path d="M9 21v-6h6v6" />
            <path d="M9 13h.01M15 13h.01M9 9h.01M15 9h.01" />
          </svg>
          </div>
          <router-link to="/login" class="flex items-center gap-2">
          <span class="auth-brand-glow auth-hero-text text-lg font-semibold tracking-wide">
            Schole
          </span>
        </router-link>
        </div>

        <div class="relative z-10 space-y-8">
          <div>
            <p class="auth-accent mb-2 text-xs font-medium uppercase tracking-[0.2em]">
              E-learning platform
            </p>
            <h2 class="auth-hero-text text-3xl font-bold leading-tight">
              Learn without
              <span class="auth-title">limits</span>
            </h2>
          </div>

          <blockquote class="auth-quote space-y-3">
            <p class="auth-hero-muted text-base leading-relaxed">
              &ldquo;Schole transformed how we manage courses, track grades,
              and review student submissions — all in one place.&rdquo;
            </p>
            <footer class="auth-accent text-sm">
              — Sofia Davis, Academic Director
            </footer>
          </blockquote>

          <div class="flex gap-4">
            <div class="auth-hero-stat flex-1 rounded-lg px-4 py-3">
              <p class="auth-hero-text text-2xl font-bold">
                Expert Instructors
              </p>
              <p class="auth-hero-muted text-xs uppercase tracking-wider">
                Learn from experienced educators
              </p>
            </div>
            <div class="auth-hero-stat flex-1 rounded-lg px-4 py-3">
              <p class="auth-hero-text text-2xl font-bold">
                Flexible Learning
              </p>
              <p class="auth-hero-muted text-xs uppercase tracking-wider">
                Study anytime, anywhere
              </p>
            </div>
          </div>
        </div>
      </div>

      <div class="flex w-full flex-col items-center justify-center">
        <div class="w-full max-w-[400px] space-y-6">
          <div class="space-y-2 text-center lg:text-left">
            <h1 class="auth-title text-3xl font-bold">
              {{ title }}
            </h1>
            <p class="auth-muted text-sm">
              {{ description }}
            </p>
          </div>

          <slot />

          <p
            v-if="showTerms !== false"
            class="text-center text-xs leading-relaxed auth-muted lg:text-left"
          >
            By continuing, you agree to our
            <a
              href="#terms"
              class="auth-link underline-offset-4 hover:underline"
            >
              Terms of Service
            </a>
            and
            <a
              href="#privacy"
              class="auth-link underline-offset-4 hover:underline"
            >
              Privacy Policy
            </a>
            .
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
