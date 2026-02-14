package harmonica

// FPS is the render frame rate used by the spring simulation.
type FPS float64

// Spring is a lightweight spring simulation compatible with bubbles/progress.
type Spring struct {
	fps       float64
	frequency float64
	damping   float64
}

// NewSpring creates a new Spring.
func NewSpring(fps FPS, frequency, damping float64) Spring {
	valueFPS := float64(fps)
	if valueFPS <= 0 {
		valueFPS = 60
	}
	if frequency <= 0 {
		frequency = 1
	}
	if damping <= 0 {
		damping = 1
	}
	return Spring{fps: valueFPS, frequency: frequency, damping: damping}
}

// Update advances the spring one frame toward target.
func (s Spring) Update(position, velocity, target float64) (float64, float64) {
	dt := 1.0 / s.fps
	stiffness := s.frequency * s.frequency
	damp := 2 * s.damping * s.frequency
	acceleration := stiffness*(target-position) - damp*velocity
	velocity += acceleration * dt
	position += velocity * dt
	return position, velocity
}
