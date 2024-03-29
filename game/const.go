package game

const SEND_STATE_PER = 2

const MAP_WIDTH = 1728
const MAP_HEIGHT = 1728
const MAP_MARGIN = 324

const V_MAX = 3.
const V_MIN = 2.
const V_ATTACK = 1000.
const V_K = 0.2
const MASS_INIT = 5 // 5
const MASS_K = 0.01
const STRENGTH_INIT = 100
const STRENGTH_COLLISION_K = 0.002
const STRENGTH_HIT_K = 0.01
const STRENGTH_HEAL = 0.01

const RADIUS_M = 50

const BULLET_V = 1.
const BULLET_LIFE = 100
const BULLET_K = 0.8
const BULLET_NEED = 1
const MAX_BULLET_RATIO = 1. / 3
const MAX_BULLET_MASS = 1000
const KB_GAIN = 1
const KB_C = 2
const COLLIDE_K = 0.995

const HITSTOP = 8
const INOPERABLE_K = 2
const INOPERABLE = 16
const CHECK_APPROACHING_EPS = 0.00001

const PRESS_RECOVER = 0.2
const PRESS_V_K = 0.5
const PRESS_REDUCE = 0.9998244353
const PRESS_REDUCE_C = 0.5

const DEAD_MASS_CENTER = 0.6
const DEAD_MASS_MINI = 0.01
const DEAD_MASS_MINI_NUM = 5
const DEAD_MASS_MINI_V = 1

const COMBAT_FRAME = 60 * 10
