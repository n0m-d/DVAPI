<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { LockKeyhole, RefreshCw, Shield } from 'lucide-vue-next'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getFounderNote } from '@/api/vault'
import { ApiError } from '@/lib/api'
import { cn } from '@/lib/utils'

const founderNote = ref('')
const loading = ref(true)
const errorMessage = ref('')
const revealed = ref(false)

const isFlag = computed(() => /^flag\{.+\}$/i.test(founderNote.value.trim()))

async function loadFounderNotes() {
  loading.value = true
  errorMessage.value = ''
  revealed.value = false

  try {
    const response = await getFounderNote()
    founderNote.value = response.data ?? ''
    requestAnimationFrame(() => {
      revealed.value = true
    })
  }
  catch (error) {
    founderNote.value = ''
    errorMessage.value = error instanceof ApiError
      ? error.message
      : error instanceof Error
        ? error.message
        : 'Failed to open the vault.'
  }
  finally {
    loading.value = false
  }
}

onMounted(loadFounderNotes)
</script>

<template>
  <LearningPageShell
    eyebrow="Admin"
    title="Vault"
    description="A sealed chamber for notes that were never meant to linger."
  >
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/admin/dashboard">
              Schole
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbPage>Vault</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <section class="vault">
      <div class="vault-chamber" aria-live="polite">
        <div class="vault-ring vault-ring-outer" aria-hidden="true" />
        <div class="vault-ring vault-ring-mid" aria-hidden="true" />
        <div class="vault-grain" aria-hidden="true" />

        <header class="vault-header">
          <div class="vault-seal" :class="{ 'vault-seal-open': revealed && !loading && !errorMessage }">
            <Shield class="vault-seal-shield" aria-hidden="true" />
            <LockKeyhole class="vault-seal-lock" aria-hidden="true" />
          </div>

          <div class="vault-meta">
            <p class="vault-kicker">Restricted archive</p>
            <h3 class="vault-heading">Founder notes</h3>
            <p class="vault-clearance">
              Clearance · Admin only
            </p>
          </div>
        </header>

        <div class="vault-divider" aria-hidden="true">
          <span />
        </div>

        <div class="vault-body">
          <div
            v-if="loading"
            class="vault-state"
          >
            <div class="vault-pulse" aria-hidden="true" />
            <p>Turning the dial…</p>
          </div>

          <div
            v-else-if="errorMessage"
            class="vault-state vault-state-error"
          >
            <p>{{ errorMessage }}</p>
            <Button
              variant="outline"
              size="sm"
              class="mt-3"
              @click="loadFounderNotes"
            >
              <RefreshCw class="mr-2 size-4" />
              Try again
            </Button>
          </div>

          <article
            v-else
            :class="cn('vault-note', revealed && 'vault-note-in', isFlag && 'vault-note-flag')"
          >
            <p class="vault-note-label">
              {{ isFlag ? 'Seal broken' : 'Sealed message' }}
            </p>
            <p class="vault-note-text">
              {{ founderNote }}
            </p>
          </article>
        </div>

        <footer class="vault-footer">
          <span>Epoch vault</span>
          <span class="vault-footer-dot" aria-hidden="true" />
          <span>Handle with care</span>
        </footer>
      </div>
    </section>
  </LearningPageShell>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500;650&family=IBM+Plex+Mono:wght@400;500&display=swap');
</style>

<style scoped>
.vault {
  --vault-brass: oklch(0.72 0.11 78);
  --vault-brass-deep: oklch(0.52 0.09 68);
  --vault-ink: color-mix(in oklch, var(--foreground) 88%, oklch(0.4 0.04 70));
  --vault-panel: color-mix(in oklch, var(--card) 78%, oklch(0.22 0.02 70));
  --vault-panel-edge: color-mix(in oklch, var(--vault-brass) 35%, var(--border));

  display: grid;
  place-items: center;
  padding-block: 0.5rem 1.5rem;
}

.vault-chamber {
  position: relative;
  width: min(100%, 40rem);
  overflow: hidden;
  border: 1px solid var(--vault-panel-edge);
  border-radius: 1.25rem;
  background:
    radial-gradient(ellipse 90% 70% at 50% 0%, color-mix(in oklch, var(--vault-brass) 14%, transparent), transparent 55%),
    radial-gradient(ellipse 70% 50% at 80% 100%, color-mix(in oklch, var(--primary) 10%, transparent), transparent 50%),
    var(--vault-panel);
  box-shadow:
    inset 0 1px 0 color-mix(in oklch, white 10%, transparent),
    0 24px 60px -28px color-mix(in oklch, var(--foreground) 28%, transparent);
  padding: 2rem 1.75rem 1.5rem;
}

.vault-ring {
  pointer-events: none;
  position: absolute;
  inset: 0;
  border-radius: inherit;
  border: 1px solid color-mix(in oklch, var(--vault-brass) 18%, transparent);
  mask-image: radial-gradient(circle at 50% 28%, transparent 18%, black 42%);
}

.vault-ring-outer {
  inset: 0.65rem;
  animation: vault-breathe 7s ease-in-out infinite;
}

.vault-ring-mid {
  inset: 1.2rem;
  border-color: color-mix(in oklch, var(--vault-brass) 10%, transparent);
  animation: vault-breathe 7s ease-in-out infinite reverse;
}

.vault-grain {
  pointer-events: none;
  position: absolute;
  inset: 0;
  opacity: 0.18;
  background-image:
    radial-gradient(circle at 20% 20%, color-mix(in oklch, var(--vault-brass) 18%, transparent) 0 1px, transparent 1.5px),
    radial-gradient(circle at 80% 60%, color-mix(in oklch, var(--foreground) 12%, transparent) 0 1px, transparent 1.5px);
  background-size: 18px 18px, 22px 22px;
  mix-blend-mode: soft-light;
}

.vault-header {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 1.1rem;
}

.vault-seal {
  position: relative;
  display: grid;
  place-items: center;
  width: 3.75rem;
  height: 3.75rem;
  flex-shrink: 0;
  border-radius: 9999px;
  border: 1px solid color-mix(in oklch, var(--vault-brass) 45%, transparent);
  background:
    radial-gradient(circle at 35% 30%, color-mix(in oklch, var(--vault-brass) 35%, transparent), transparent 55%),
    color-mix(in oklch, var(--background) 55%, transparent);
  color: var(--vault-brass-deep);
  box-shadow: inset 0 0 0 4px color-mix(in oklch, var(--vault-brass) 12%, transparent);
  transition: transform 0.55s ease, border-color 0.55s ease, color 0.55s ease;
}

.vault-seal-open {
  transform: rotate(-8deg) scale(1.04);
  border-color: color-mix(in oklch, var(--vault-brass) 70%, transparent);
  color: var(--vault-brass);
}

.vault-seal-shield {
  width: 1.65rem;
  height: 1.65rem;
  opacity: 0.9;
}

.vault-seal-lock {
  position: absolute;
  width: 0.85rem;
  height: 0.85rem;
  bottom: 0.7rem;
  opacity: 0.85;
}

.vault-meta {
  min-width: 0;
}

.vault-kicker,
.vault-clearance,
.vault-note-label,
.vault-footer {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.vault-kicker {
  margin: 0;
  font-size: 0.65rem;
  color: var(--vault-brass-deep);
}

.vault-heading {
  margin: 0.2rem 0 0.35rem;
  font-family: Fraunces, ui-serif, Georgia, serif;
  font-size: clamp(1.6rem, 3vw, 2rem);
  font-weight: 650;
  letter-spacing: -0.02em;
  color: var(--vault-ink);
  line-height: 1.1;
}

.vault-clearance {
  margin: 0;
  font-size: 0.62rem;
  color: var(--muted-foreground);
}

.vault-divider {
  position: relative;
  z-index: 1;
  display: grid;
  place-items: center;
  margin: 1.5rem 0 1.25rem;
}

.vault-divider span {
  display: block;
  width: min(100%, 18rem);
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in oklch, var(--vault-brass) 55%, transparent),
    transparent
  );
  animation: vault-pulse-line 4.5s ease-in-out infinite;
}

.vault-body {
  position: relative;
  z-index: 1;
  min-height: 9.5rem;
  display: grid;
  place-items: center;
}

.vault-state {
  display: grid;
  place-items: center;
  gap: 0.75rem;
  text-align: center;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 0.8rem;
  letter-spacing: 0.04em;
}

.vault-state-error {
  color: var(--destructive);
}

.vault-pulse {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 9999px;
  border: 2px solid color-mix(in oklch, var(--vault-brass) 35%, transparent);
  border-top-color: var(--vault-brass);
  animation: vault-spin 1s linear infinite;
}

.vault-note {
  width: 100%;
  opacity: 0;
  transform: translateY(10px);
  transition: opacity 0.55s ease, transform 0.55s ease;
  text-align: center;
  padding: 0.25rem 0.5rem 0.5rem;
}

.vault-note-in {
  opacity: 1;
  transform: translateY(0);
}

.vault-note-label {
  margin: 0 0 0.85rem;
  font-size: 0.62rem;
  color: color-mix(in oklch, var(--vault-brass-deep) 80%, var(--muted-foreground));
}

.vault-note-text {
  margin: 0;
  font-family: Fraunces, ui-serif, Georgia, serif;
  font-size: clamp(1.25rem, 2.8vw, 1.65rem);
  font-weight: 500;
  line-height: 1.45;
  color: var(--vault-ink);
  text-wrap: balance;
  white-space: pre-wrap;
}

.vault-note-flag .vault-note-text {
  color: var(--vault-brass-deep);
  letter-spacing: 0.01em;
}

.vault-note-flag .vault-note-label {
  color: var(--vault-brass);
}

.vault-footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  margin-top: 1.75rem;
  font-size: 0.6rem;
  color: color-mix(in oklch, var(--muted-foreground) 85%, var(--vault-brass));
}

.vault-footer-dot {
  width: 0.28rem;
  height: 0.28rem;
  border-radius: 9999px;
  background: color-mix(in oklch, var(--vault-brass) 70%, transparent);
  animation: vault-dot 2.8s ease-in-out infinite;
}

@keyframes vault-breathe {
  0%,
  100% {
    opacity: 0.45;
    transform: scale(1);
  }

  50% {
    opacity: 0.9;
    transform: scale(1.01);
  }
}

@keyframes vault-pulse-line {
  0%,
  100% {
    opacity: 0.35;
  }

  50% {
    opacity: 1;
  }
}

@keyframes vault-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes vault-dot {
  0%,
  100% {
    opacity: 0.35;
    transform: scale(0.9);
  }

  50% {
    opacity: 1;
    transform: scale(1.15);
  }
}

@media (max-width: 640px) {
  .vault-chamber {
    padding: 1.5rem 1.15rem 1.25rem;
  }

  .vault-header {
    align-items: flex-start;
  }
}
</style>
