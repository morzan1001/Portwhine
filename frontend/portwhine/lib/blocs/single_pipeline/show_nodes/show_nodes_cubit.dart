import 'package:bloc/bloc.dart';

class ShowNodesCubit extends Cubit<bool> {
  ShowNodesCubit() : super(false);

  void toggleNodes() => emit(!state);
}
