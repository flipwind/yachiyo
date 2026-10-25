class RuntimeState {
  final DateTime timetick;
  final double energy;
  final RuntimeAffect affect;

  RuntimeState({required this.timetick, required this.energy, required this.affect});
}

class RuntimeAffect{
  final String octants;
  final double pleasure;
  final double arousal;
  final double dominance;
  final List<OCCEmotion> actives;

  RuntimeAffect({required this.octants, required this.pleasure, required this.arousal, required this.dominance, required this.actives});
}

class OCCEmotion {
  final String type;
  final double intensity;

  OCCEmotion({required this.type, required this.intensity});
}