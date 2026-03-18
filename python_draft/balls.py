# /// script
# dependencies = [
#   "pygame-ce",
# ]
# ///

import json
import os
import sys
import time

import pygame

# Key Controls Reminder:
# Shift + 1..5: Save current pattern to file and memory slot.
# 1..5: Load pattern from slot.
# A: Toggle Auto-animation mode (cycles through saved slots).
# C: Clear all balls (smoothly).
# Mouse Click: Toggle cell state manually.

# ==========================================
# --- SETTINGS (CONSTANTS) ---
# ==========================================
COLS, ROWS = 9, 21  # Grid dimensions
GRID_SUB = 4  # Sub-grid division factor

# Timing and Speeds
ANIM_SPEED = 0.08  # Ball movement speed (0.01 - slow, 0.5 - instant)
SLIDE_TIME = 4.0  # Pause duration between pattern changes (seconds)
SSAA_SCALE = 2.0  # Anti-aliasing quality (1.0 - off, 2.0 - standard, 4.0 - max)

# Color Palette
COLOR_BG_BLUE = (63, 81, 151)
COLOR_BG_LIGHT = (238, 234, 214)
COLOR_GRID_GRAY = (160, 160, 160)
BALL_DARK = (26, 31, 74)
BALL_WHITE = (220, 225, 235)
BALL_GLARE = (255, 255, 255)
BALL_OUTLINE = (20, 20, 40)
# ==========================================


class Ball:
    def __init__(self, color):
        self.x, self.y = 0, 0
        self.tx, self.ty = 0, 0
        self.color = color
        self.alpha = 0
        self.radius = 6

    def update(self, target_x, target_y, visible):
        self.tx, self.ty = target_x, target_y
        # Smooth interpolation for movement
        self.x += (self.tx - self.x) * ANIM_SPEED
        self.y += (self.ty - self.y) * ANIM_SPEED

        # Alpha transition for fading in/out
        target_alpha = 255 if visible else 0
        self.alpha += (target_alpha - self.alpha) * ANIM_SPEED

    def draw(self, surface, scale):
        if self.alpha < 5:
            return
        r = int(self.radius * scale)
        pos = (int(self.x * scale), int(self.y * scale))

        # Main body
        pygame.draw.circle(surface, self.color, pos, r)
        # Outline
        pygame.draw.circle(surface, BALL_OUTLINE, pos, r, max(1, int(scale)))
        # Glare/Highlight
        glare_pos = (pos[0] - r // 3, pos[1] - r // 3)
        pygame.draw.circle(surface, BALL_GLARE, glare_pos, r // 4)


class Cell:
    def __init__(self, c, r):
        self.c, self.r = c, r
        self.state = 0
        is_blue = (c + r) % 2 == 0
        self.balls = [Ball(BALL_WHITE if is_blue else BALL_DARK) for _ in range(4)]
        self.initialized = False

    def update_logic(self, cw, ch):
        bx, by = self.c * cw, self.r * ch
        mid_x, mid_y = bx + cw / 2, by + ch / 2

        if not self.initialized:
            for b in self.balls:
                b.x, b.y = mid_x, mid_y
            self.initialized = True

        sw, sh = cw / GRID_SUB, ch / GRID_SUB
        # Target points for the 4 balls in a cell
        pts = [
            (bx + sw / 2, by + sh / 2),
            (bx + cw - sw / 2, by + ch - sh / 2),
            (bx + cw - sw / 2, by + sh / 2),
            (bx + sw / 2, by + ch - sh / 2),
        ]

        for i, b in enumerate(self.balls):
            # Logic for which balls are visible based on cell state
            is_visible = (
                (self.state == 1 and i < 2)
                or (self.state == 2 and i >= 2)
                or (self.state == 3)
            )

            target = pts[i] if is_visible else (mid_x, mid_y)
            b.update(target[0], target[1], is_visible)


def main():
    pygame.init()
    win_w, win_h = 450, 850
    screen = pygame.display.set_mode((win_w, win_h), pygame.RESIZABLE)
    pygame.display.set_caption("Pattern Studio")
    clock = pygame.time.Clock()

    cells = [[Cell(c, r) for c in range(COLS)] for r in range(ROWS)]
    slots = {i: None for i in range(1, 6)}

    # Initial file loading for saved patterns
    for i in range(1, 6):
        file_path = f"slot_{i}.json"
        if os.path.exists(file_path):
            with open(file_path, "r") as f:
                slots[i] = json.load(f)

    auto_mode = False
    last_switch_time = time.time()
    current_slot_index = 0

    def apply_pattern(data):
        if not data:
            return
        for r in range(ROWS):
            for c in range(COLS):
                if r < len(data) and c < len(data[0]):
                    cells[r][c].state = data[r][c]

    while True:
        # Filter non-empty slots for the animation playlist
        active_slots = [s for s in slots.values() if s is not None]

        # Auto-animation logic
        if auto_mode and active_slots and time.time() - last_switch_time > SLIDE_TIME:
            current_slot_index = (current_slot_index + 1) % len(active_slots)
            apply_pattern(active_slots[current_slot_index])
            last_switch_time = time.time()

        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                pygame.quit()
                sys.exit()

            if event.type == pygame.KEYDOWN:
                # Clear board
                if event.key == pygame.K_c:
                    for r in range(ROWS):
                        for c in range(COLS):
                            cells[r][c].state = 0

                # Toggle auto-mode
                if event.key == pygame.K_a:
                    auto_mode = not auto_mode
                    last_switch_time = time.time()

                # Slot handling (1-5)
                if pygame.K_1 <= event.key <= pygame.K_5:
                    slot_num = event.key - pygame.K_0
                    # Shift + Number = Save
                    if pygame.key.get_mods() & pygame.KMOD_SHIFT:
                        current_data = [
                            [cells[r][c].state for c in range(COLS)]
                            for r in range(ROWS)
                        ]
                        slots[slot_num] = current_data
                        with open(f"slot_{slot_num}.json", "w") as f:
                            json.dump(current_data, f)
                    # Number only = Load
                    else:
                        if slots[slot_num]:
                            apply_pattern(slots[slot_num])
                            auto_mode = False

            if event.type == pygame.MOUSEBUTTONDOWN:
                # Calculate coordinates based on current window size
                curr_w, curr_h = screen.get_size()
                cw_calc, ch_calc = curr_w / COLS, curr_h / ROWS
                mx, my = pygame.mouse.get_pos()
                c_idx, r_idx = int(mx // cw_calc), int(my // ch_calc)

                if 0 <= c_idx < COLS and 0 <= r_idx < ROWS:
                    cells[r_idx][c_idx].state = (cells[r_idx][c_idx].state + 1) % 4
                    auto_mode = False

        # --- RENDERING PIPELINE (SSAA) ---
        win_w, win_h = screen.get_size()
        cw, ch = win_w / COLS, win_h / ROWS

        # Create a high-resolution surface for anti-aliasing
        virt_w, virt_h = int(win_w * SSAA_SCALE), int(win_h * SSAA_SCALE)
        canvas = pygame.Surface((virt_w, virt_h))
        canvas.fill(COLOR_BG_LIGHT)

        # Draw Background (checkerboard)
        for r in range(ROWS):
            for c in range(COLS):
                if (r + c) % 2 == 0:
                    rect = (
                        int(c * cw * SSAA_SCALE),
                        int(r * ch * SSAA_SCALE),
                        int(cw * SSAA_SCALE) + 1,
                        int(ch * SSAA_SCALE) + 1,
                    )
                    pygame.draw.rect(canvas, COLOR_BG_BLUE, rect)

        # Draw Grid Lines
        sw, sh = cw / GRID_SUB * SSAA_SCALE, ch / GRID_SUB * SSAA_SCALE
        for i in range(COLS * GRID_SUB + 1):
            pygame.draw.line(
                canvas,
                COLOR_GRID_GRAY,
                (int(i * sw), 0),
                (int(i * sw), virt_h),
                int(SSAA_SCALE),
            )
        for i in range(ROWS * GRID_SUB + 1):
            pygame.draw.line(
                canvas,
                COLOR_GRID_GRAY,
                (0, int(i * sh)),
                (virt_w, int(i * sh)),
                int(SSAA_SCALE),
            )

        # Update and Draw Balls
        for r in range(ROWS):
            for c in range(COLS):
                cells[r][c].update_logic(cw, ch)
                for ball in cells[r][c].balls:
                    ball.draw(canvas, SSAA_SCALE)

        # Scale down the high-res canvas to the window size (SSAA effect)
        final_frame = pygame.transform.smoothscale(canvas, (win_w, win_h))
        screen.blit(final_frame, (0, 0))

        pygame.display.flip()
        clock.tick(60)


if __name__ == "__main__":
    main()
