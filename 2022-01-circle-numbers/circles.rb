class Circle

  # PREPOPULATED ARRAYS OF PRIMES WITH n DIGITS
  PRIMES_WITH_UNIQ_DIGITS = Hash.new do |h, k|
    h[k] = Marshal.load(File.binread("primes#{k}")).select do |i|
      digits = i.digits
      digits.size == digits.uniq.size
    end
  end

  def initialize(*digits, d:)
    @digits = digits
    @n = digits.size
    @d = d
  end

  def circle_score
    @circle_score ||= PRIMES_WITH_UNIQ_DIGITS[@d].sum do |i|
      (i.digits - @digits).empty? ? number_score(i) : 0
    end
  end

  def number_score(i)
    i.digits.each_cons(2).sum do |a, b|
      ai = @digits.index(a)
      bi = @digits.index(b)
      [(ai - bi) % @n, (bi - ai) % @n].min
    end
  end

  def <=>(other)
    circle_score <=> other.circle_score
  end

  def to_s
    @digits.inspect
  end

end

class Circles

  include Enumerable

  def initialize(n, d)
    @n, @d = n, d
  end

  def each
    (0..9).to_a.permutation(@n) do |a|
      # Reduce solution space by removing symmetric circles
      # * Cyclic symmetry  - only consider circles with first minimum first digit
      # * Reflective symmetry -  only consider circles with second digit greater than next to last
      next unless a[0] == a.min && a[1] > a[-1]
      yield Circle.new(*a, d: @d)
    end
  end

  def minmax
    circles = Parallel.map(self, in_processes: 16, progress: true) do |c|
      c.tap(&:circle_score)
    end.sort.minmax
  end

end

puts Circles.new(7, 5).minmax
puts Circles.new(8, 6).minmax
