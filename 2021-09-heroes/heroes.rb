class H

  attr_reader :villains, :total_damage

  def initialize(villains, total_damage)
    @villains = villains.chars.map(&:to_i)
    @total_damage = total_damage
  end

  def damage_per_villain
    @damage_per_villain ||= total_damage.to_f / villains.size
  end

  def scan(pattern, fudge_factor: 1)
    damage_sum = 0
    villains.size.times.map do |i|
      match(pattern, fudge_factor: fudge_factor, i_v_offset: i)
    end.compact.sort_by(&:last).each do |i, v, h, damage|

    
  end

  def match(pattern, fudge_factor: 1, i_v_offset: 0)
    heroes = pattern.chars.map(&:to_i)

    damage = 0
    v =  ''
    h = ''
    i_v = i_v_offset
    i_h = 0
    while i_v < villains.length && i_h < heroes.length
      d1 = abs(villains[i_v] - heroes[i_h])
      d2 = abs(villains[i_v])
      d3 = abs(heroes[i_h])
      if d1 <= d2 && d1 <= d3
        v << villains[i_v].to_s
        h << heroes[i_h].to_s
        i_v += 1
        i_h += 1
        damage += d1
      elsif d2 <= d3
        h << '-' # hyphen heroes
        v << villains[i_v].to_s
        i_v += 1
        damage += d2
      else
        v << '-' # hyphen villain
        h << heroes[i_h].to_s
        i_h += 1
        damage += d3
      end
      return if damage > damage_per_villain * (i_v - i_v_offset + heroes.length - i_h + fudge_factor)
    end
    [i_v_offset, v, h, damage] if i_h >= heroes.length
  end

  def abs(i)
    i >= 0 ? i : -i
  end

end

s = '31415926535897932384626433832795028841971693993751058209749445923078164'
p s[39..48]
#s = '31415949335897932384626433832795028841971693993751058209749445923078164'
h = H.new(s, 50)
h.scan('9793')
#p h.match('9793', 1, 0)
#p h.match('9793', 1, 5)
