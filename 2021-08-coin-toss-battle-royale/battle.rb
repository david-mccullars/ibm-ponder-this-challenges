class BattleCycle

  def initialize(players:, inbound_prob:, coins:)
    @players = players
    @inbound_prob = inbound_prob
    raise "Inbound probabilities must sum to <= 1.0" if inbound_prob.sum > 1.0
    @coins = coins
  end

  def size
    @size ||= @players.size
  end

  def victor
    @players.first if size == 1
  end

  def in_cycle_prob
    @in_cycle_prob ||= @players.map { |player| 1.0 - @coins[player] }
  end

  def infinite_multiplier
    @infinite_multiplier ||= (1.0 / (1.0 - in_cycle_prob.reduce(&:*))) # Infinite sum of geometric series
  end

  def nodes
    @nodes ||= size.times.map do |i|
      @players.rotate(i)
    end
  end

  def outbound_nodes
    @outbound_nodes ||= nodes.map do |node|
      biggest_threat = player_threats[node.first]
      node.rotate(1) - [biggest_threat]
    end
  end

  def outbound_prob
    @outbound_prob ||= size.times.map do |j|
      size.times.map do |i|
        i2j_prob = in_cycle_prob.rotate(i)[0 ... (size + j - i) % size].reduce(&:*) || 1.0
        @inbound_prob[i] * i2j_prob * @coins[@players[j]] * infinite_multiplier
      end.reduce(&:+)
    end
  end

  def outbound_cycles
    outbound_nodes.zip(outbound_prob).group_by { |n, p| n.sort }.map do |cycle, inbound|
      BattleCycle.new(
        players: cycle,
        inbound_prob: cycle.map { |player| inbound.detect { |n, p| n[0] == player }&.last || 0.0 },
        coins: @coins,
      )
    end
  end

  def victors
    @victors ||= Hash.new { 0 }.tap do |sums|
      if size == 1
        sums[@players.first] = @inbound_prob.first
      else
        outbound_cycles.each do |cycle|
          cycle.victors.each do |player, prob|
            sums[player] += prob
          end
        end
      end
    end
  end

  def player_threats
    @player_threats ||= @players.each_with_index.map do |player, position|
      target_offset_most_concerning = 1.upto(size - 1).max_by do |target_offset|
        # Have to find the probability (after winning and removing target) to
        # receive another turn without someone else winning first
        rotated = in_cycle_prob.rotate(position + 1)[0 ... -1]
        rotated.delete_at(target_offset - 1)
        prob = rotated.reduce(&:*) || 1.0
        [@coins[player] * prob, -target_offset]
      end
      [player, @players[(position + target_offset_most_concerning) % size]]
    end.to_h
  end

end

class BattleGraph

  def initialize(*coin_prob)
    @coins = coin_prob.each_with_index.to_a.map(&:reverse).to_h
    @cycle = BattleCycle.new(
      players: @coins.keys,
      inbound_prob: [1.0] + Array.new(@coins.size - 1) { 0 },
      coins: @coins,
    )
p @cycle
p @cycle.in_cycle_prob
p @cycle.player_threats
p @cycle.outbound_nodes
puts "===="
c2 = @cycle.outbound_cycles.first
p c2
p c2.in_cycle_prob
p c2.player_threats
p c2.outbound_nodes
  end

  def victors
    @victors ||= @cycle.victors.sort.to_h
  end

  def sigma_c
    @sigma_c ||= sigma_for(@coins)
  end

  def sigma_w
    @sigma_w ||= sigma_for(victors)
  end

  def sigma_for(h)
    sorted = h.sort_by(&:last).map(&:first)
    h.keys.map { |k| sorted.index(k) + 1 }
  end

  def fair?
    sigma_c == sigma_w
  end

end

puts "==========="
b = BattleGraph.new(0.25, 0.5, 1.0)
p b.victors
p b.sigma_c
p b.sigma_w
p b.fair?
puts "==========="
b = BattleGraph.new(0.5, 0.2, 0.05, 0.85, 0.1)
__END__
p b.victors
p b.sigma_c
p b.sigma_w
p b.fair?
puts "==========="

b = BattleGraph.new(0.035, 0.008, 0.5, 0.9, 0.25, 0.017, 0.125, 0.07)
p b.victors
p b.sigma_c
p b.sigma_w
p b.fair?



__END__

winners = Hash.new { 0 }
#coins = { 0 => 0.25, 1 => 0.5, 2 => 1.0 }
coins = { 0 => 0.5, 1 => 0.2, 2 => 0.05, 3 => 0.85, 4 => 0.1 }


100000000.times do
  players = coins.keys.sort
  loop do
    if rand <= coins[players.first]
      biggest_threat = players[1..].each_with_index.max_by { |p, x| [coins[p], -x] }.first
      players.delete(biggest_threat)
    end
    break if players.size == 1
    players = players.rotate(1)
  end
  winners[players[0]] += 1
end
p winners.sort.to_h
