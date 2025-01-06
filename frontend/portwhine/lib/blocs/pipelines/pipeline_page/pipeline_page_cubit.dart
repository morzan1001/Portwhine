import 'package:bloc/bloc.dart';

class PipelinePageCubit extends Cubit<int> {
  PipelinePageCubit() : super(1);

  void setFirstPage() => emit(1);
  void nextPage() => emit(state + 1);

  void previousPage() {
    if (state == 1) return;
    emit(state - 1);
  }
}
