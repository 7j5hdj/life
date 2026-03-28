package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 480
	screenHeight = 480
	cellSize     = 4
	uiHeight     = 40
)

type Game struct {
	world      [][]bool      //Мир из 0 и 1 для логики
	width      int           //Ширина
	height     int           //Высота
	ticks      int           //Тики "FPS"
	fillRate   float64       //Текущий процент заполнения (0.0 - 1.0)
	cellImg    *ebiten.Image //Кэшируем изображение клетки для оптимизации
	generation int           //Счётчик поколений
	paused     bool          //Пауза
}

func NewGame() *Game {
	h, w := screenWidth/cellSize, screenHeight/cellSize
	//создаем текстуру клетки один раз при старте
	img := ebiten.NewImage(cellSize-1, cellSize-1)

	img.Fill(color.RGBA{0, 255, 127, 255})

	g := &Game{
		width:    w,
		height:   h,
		fillRate: 0.2, // По умолчанию 20%
		cellImg:  img,
	}
	g.reset(g.fillRate)
	return g
}

// reset пересоздает мир с заданным процентом заполнения
func (g *Game) reset(rate float64) {
	g.generation = 0
	g.fillRate = rate
	g.world = make([][]bool, g.height)
	for i := range g.world {
		g.world[i] = make([]bool, g.width)
		for j := range g.world[i] {
			g.world[i][j] = rand.Float64() < g.fillRate
		}
	}
}

func (g *Game) Update() error {
	// 1. Обработка ввода для смены процента
	for i := 1; i <= 9; i++ {
		// ebiten.Key1, Key2... это константы, мы можем их проверить
		if ebiten.IsKeyPressed(ebiten.Key0 + ebiten.Key(i)) {
			g.reset(float64(i) * 0.1)
		}
	}

	// Клавиша R для простого рестарта с тем же процентом
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.reset(g.fillRate)
	}
	//Кавиша Space для остановки симуляции
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	// 2. Логика симуляции
	g.ticks++
	if g.ticks%5 != 0 {
		return nil
	}
	//Сначала создаём копию мира записываем туда изменения и в конце производим замену старого мира на новый

	newWorld := make([][]bool, g.height)
	for y := 0; y < g.height; y++ {

		newWorld[y] = make([]bool, g.width)
		for x := 0; x < g.width; x++ {

			neighbors := g.countNeighbors(x, y)
			if g.world[y][x] && (neighbors == 2 || neighbors == 3) { //Клетка живёт при 2 или 3 соседях(Выживание)
				newWorld[y][x] = true
			} else if !g.world[y][x] && neighbors == 3 { //Клетка мертва при 3 соседях(Перенаселение-Смерть)
				newWorld[y][x] = true
			}

		}
	}
	if !g.paused {
		g.world = newWorld //Замена
		g.generation++     // Счётчик поколений
	}

	return nil
}

func (g *Game) countNeighbors(x, y int) int { //Окресность мура подсчёт живых клеток
	count := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			nx, ny := (x+i+g.width)%g.width, (y+j+g.height)%g.height //Тороидальный мир
			if g.world[ny][nx] {
				count++
			}
		}
	}
	return count
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 10, 255})

	opts := &ebiten.DrawImageOptions{}
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			if g.world[y][x] {
				// Сбрасываем и настраиваем матрицу для каждой клетки
				opts.GeoM.Reset()
				opts.GeoM.Translate(float64(x*cellSize), float64(y*cellSize+uiHeight))
				screen.DrawImage(g.cellImg, opts)
			}
		}
	}

	// Вывод информации
	msg := fmt.Sprintf("Fill percentage %d%% | Generation counter %d | R:Reset | Space:Stop", int(g.fillRate*100), g.generation)

	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(w, h int) (int, int) { return screenWidth, screenHeight }

func main() {
	ebiten.SetWindowTitle("LifeMur")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
