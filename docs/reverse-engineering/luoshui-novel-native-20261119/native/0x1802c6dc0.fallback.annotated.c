/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c6dc0 */

longlong * FUN_1802c6dc0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  int *piVar3;
  char cVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  undefined8 uVar10;
  code *pcVar11;
  uint uVar12;
  longlong lVar13;
  longlong *plVar14;
  longlong *plVar15;
  undefined8 uVar16;
  code *pcVar17;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_60;
  
  local_res18 = DAT_1803ab250;
  plVar1 = (longlong *)param_3[1];
  plVar2 = (longlong *)*param_3;
  local_60 = (longlong *)0x0;
  iVar5 = (int)*DAT_1803ab250 + 1;
  if (iVar5 != 0) {
    *(int *)DAT_1803ab250 = iVar5;
  }
  pcVar17 = _Py_FalseStruct_exref;
  if (*(int *)_Py_FalseStruct_exref + 1 != 0) {
    *(int *)_Py_FalseStruct_exref = *(int *)_Py_FalseStruct_exref + 1;
  }
  plVar15 = DAT_1803ab250;
  local_res20 = DAT_1803ab250;
  iVar5 = (int)*DAT_1803ab250 + 1;
  if (iVar5 != 0) {
    *(int *)DAT_1803ab250 = iVar5;
  }
  plVar14 = DAT_1803c4bb0;
  if (DAT_1803c4bb0 == (longlong *)0x0) {
LAB_1802c6e6c:
    DAT_1803c4bb0 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4b50,DAT_1803ca468,0x30);
  }
  else if ((1 < *DAT_1803c4bb0) || (DAT_1803c4bb0[2] != 0)) {
    if ((DAT_1803c4bb0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4bb0 && (*DAT_1803c4bb0 = *DAT_1803c4bb0 + -1, *plVar14 == 0)))) {
      (**(code **)(plVar14[1] + 0x30))();
    }
    goto LAB_1802c6e6c;
  }
  plVar14 = DAT_1803c4bb0;
  FUN_180001d00(param_1,DAT_1803c4bb0);
  uVar16 = 0x71;
  do {
    lVar6 = FUN_180373860(plVar2);
    if (((lVar6 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) ||
       (plVar7 = (longlong *)PyLong_FromSsize_t(lVar6), plVar7 == (longlong *)0x0))
    goto LAB_1802c795f;
    cVar4 = FUN_1803847a0(plVar15,plVar7);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    pcVar9 = _Py_FalseStruct_exref;
    if (cVar4 != '\0') {
      pcVar9 = _Py_TrueStruct_exref;
    }
    iVar5 = FUN_180001510(pcVar9);
    plVar15 = local_res20;
    if (iVar5 == 1) {
      plVar7 = (longlong *)FUN_180001b80(param_1,plVar2,local_res20);
      if (plVar7 != (longlong *)0x0) {
        lVar6 = plVar7[1];
        if (*(code **)(lVar6 + 0x90) == (code *)0x0) {
/* novels.services.spot_fix_patches#39="isspace" */
          uVar12 = *(uint *)(DAT_1803c4d68 + 0x20) & 0x60;
          if (*(code **)(lVar6 + 0x40) == (code *)0x0) {
/* novels.services.spot_fix_patches#39="isspace" */
            lVar13 = DAT_1803c4d68 + 0x28;
            if (uVar12 != 0x60) {
/* novels.services.spot_fix_patches#39="isspace" */
              lVar13 = *(longlong *)(DAT_1803c4d68 + 0x30);
            }
            PyErr_Format(*(undefined8 *)PyExc_AttributeError_exref,
                         "\'%s\' object has no attribute \'%s\'",*(undefined8 *)(lVar6 + 0x18),
                         lVar13);
            plVar8 = (longlong *)0x0;
          }
          else {
/* novels.services.spot_fix_patches#39="isspace" */
            lVar13 = DAT_1803c4d68 + 0x28;
            if (uVar12 != 0x60) {
/* novels.services.spot_fix_patches#39="isspace" */
              lVar13 = *(longlong *)(DAT_1803c4d68 + 0x30);
            }
            plVar8 = (longlong *)(**(code **)(lVar6 + 0x40))(plVar7,lVar13);
          }
        }
        else {
/* novels.services.spot_fix_patches#39="isspace" */
          plVar8 = (longlong *)(**(code **)(lVar6 + 0x90))(plVar7,DAT_1803c4d68);
        }
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (plVar8 != (longlong *)0x0) {
          *(undefined4 *)(plVar14 + 5) = 0x71;
          pcVar9 = (code *)FUN_180376210(param_1,plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (pcVar9 != (code *)0x0) goto LAB_1802c7032;
        }
        goto LAB_1802c7180;
      }
      goto LAB_1802c795f;
    }
    if (*(int *)pcVar9 + 1 != 0) {
      *(int *)pcVar9 = *(int *)pcVar9 + 1;
    }
LAB_1802c7032:
    iVar5 = FUN_180001510(pcVar9);
    if ((-1 < *(int *)pcVar9) &&
       (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
      (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
    }
    plVar15 = local_res20;
    if (iVar5 == -1) goto LAB_1802c795f;
    if (iVar5 == 0) {
      uVar16 = 0x77;
      goto LAB_1802c71a0;
    }
    FUN_180382bf0(&local_res20);
    plVar15 = local_res20;
    lVar6 = *(longlong *)(param_1 + 0x10);
    if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar6 + 0x1a0) != 0)) &&
        (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) && (*(longlong *)(param_1 + 0x60) != 0)) {
LAB_1802c7180:
      lVar6 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c7969;
    }
    if (*(int *)(lVar6 + 0x174) != 0) {
      PyEval_SaveThread();
      PyEval_AcquireThread(param_1);
    }
    if (*(longlong *)(param_1 + 0x80) != 0) {
      plVar7 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x80) = 0;
      uVar10 = FUN_180373490(param_1);
      *(undefined8 *)(param_1 + 0x60) = uVar10;
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      goto LAB_1802c7180;
    }
  } while (plVar15 != (longlong *)0x0);
  plVar7 = (longlong *)
           PyUnicode_FromFormat
                     ("cannot access local variable \'%U\' where it is not associated with a value",
/* novels.services.spot_fix_patches#38="oi" */
                      DAT_1803c4d60);
  lVar6 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar7);
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  FUN_18000d1d0(param_1,lVar6);
  goto LAB_1802c7969;
LAB_1802c7550:
  plVar7 = (longlong *)
           PyUnicode_FromFormat
                     ("cannot access local variable \'%U\' where it is not associated with a value",
                      uVar10);
  lVar6 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar7);
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  FUN_18000d1d0(param_1,lVar6);
  uVar16 = 0x73;
  pcVar17 = pcVar11;
  goto LAB_1802c7969;
LAB_1802c71a0:
  lVar6 = FUN_180373860(plVar2);
  if (((lVar6 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) ||
     (plVar7 = (longlong *)PyLong_FromSsize_t(lVar6), plVar7 == (longlong *)0x0)) {
    uVar16 = 0x73;
    goto LAB_1802c795f;
  }
  cVar4 = FUN_180384160(plVar15,plVar7);
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  pcVar9 = _Py_FalseStruct_exref;
  if (cVar4 != '\0') {
    pcVar9 = _Py_TrueStruct_exref;
  }
  iVar5 = FUN_180001510(pcVar9);
  pcVar11 = pcVar17;
  if (iVar5 == 1) {
/* novels.services.spot_fix_patches#40="ni" */
    uVar10 = DAT_1803c4d70;
    if (local_res18 == (longlong *)0x0) goto LAB_1802c7550;
    pcVar9 = (code *)FUN_180384550(local_res18,plVar1);
    if (pcVar9 == (code *)0x0) {
      lVar6 = *(longlong *)(param_1 + 0x60);
      uVar16 = 0x73;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c7969;
    }
  }
  else if (*(int *)pcVar9 + 1 != 0) {
    *(int *)pcVar9 = *(int *)pcVar9 + 1;
  }
  iVar5 = FUN_180001510(pcVar9);
  if ((-1 < *(int *)pcVar9) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
  if (iVar5 == -1) goto LAB_1802c76f2;
  if (iVar5 != 0) {
    lVar6 = FUN_180373860(plVar2);
    if (((lVar6 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) ||
       (plVar7 = (longlong *)PyLong_FromSsize_t(lVar6), plVar7 == (longlong *)0x0)) {
      uVar16 = 0x74;
      goto LAB_1802c795f;
    }
    cVar4 = FUN_1803847a0(plVar15,plVar7);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    plVar15 = local_res20;
    if (cVar4 == '\0') goto LAB_1802c780c;
    plVar7 = (longlong *)FUN_180001b80(param_1,plVar2,local_res20);
    if (plVar7 == (longlong *)0x0) {
      uVar16 = 0x76;
      goto LAB_1802c795f;
    }
    if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
       (*local_60 = *local_60 + -1, *local_60 == 0)) {
      (**(code **)(local_60[1] + 0x30))(local_60);
    }
    plVar8 = (longlong *)FUN_18037f0c0();
    local_60 = plVar7;
    if (plVar8 == (longlong *)0x0) {
LAB_1802c77c9:
      lVar6 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c7969;
    }
    *(undefined4 *)(plVar14 + 5) = 0x77;
    plVar7 = (longlong *)FUN_180376210(param_1,plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (plVar7 == (longlong *)0x0) goto LAB_1802c77c9;
    iVar5 = FUN_180001510(plVar7);
    if (iVar5 == -1) {
      lVar6 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      goto LAB_1802c7969;
    }
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (iVar5 == 0) {
      if (local_res18 == (longlong *)0x0) {
        plVar7 = (longlong *)
                 PyUnicode_FromFormat
                           ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#40="ni" */
                            ,DAT_1803c4d70);
        lVar6 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar7);
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        FUN_18000d1d0(param_1,lVar6);
        uVar16 = 0x7c;
        goto LAB_1802c7969;
      }
      local_res20 = local_res18;
      cVar4 = FUN_180382990(&local_res20,DAT_1803ab258);
      pcVar11 = _Py_FalseStruct_exref;
      if (cVar4 == '\0') {
        lVar6 = *(longlong *)(param_1 + 0x60);
        uVar16 = 0x7c;
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802c7969;
      }
      local_res18 = local_res20;
      if (*(int *)_Py_FalseStruct_exref + 1 != 0) {
        *(int *)_Py_FalseStruct_exref = *(int *)_Py_FalseStruct_exref + 1;
      }
      if (pcVar17 != (code *)0x0) goto LAB_1802c74ab;
    }
    else {
      if (pcVar17 == (code *)0x0) {
        plVar7 = (longlong *)
                 PyUnicode_FromFormat
                           ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#41="in_whitespace" */
                            ,DAT_1803c4d78);
        lVar6 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar7);
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        FUN_18000d1d0(param_1,lVar6);
        uVar16 = 0x78;
        goto LAB_1802c7969;
      }
      iVar5 = FUN_180001510(pcVar17);
      if (iVar5 == -1) {
        lVar6 = *(longlong *)(param_1 + 0x60);
        uVar16 = 0x78;
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802c7969;
      }
      if (iVar5 == 0) {
        if (local_res18 == (longlong *)0x0) {
          plVar7 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#40="ni" */
                              ,DAT_1803c4d70);
          lVar6 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar7);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          FUN_18000d1d0(param_1,lVar6);
          uVar16 = 0x79;
          goto LAB_1802c7969;
        }
        local_res20 = local_res18;
        cVar4 = FUN_180382990(&local_res20,DAT_1803ab258);
        pcVar11 = _Py_TrueStruct_exref;
        if (cVar4 == '\0') {
          lVar6 = *(longlong *)(param_1 + 0x60);
          uVar16 = 0x79;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1802c7969;
        }
        if (*(int *)_Py_TrueStruct_exref + 1 != 0) {
          *(int *)_Py_TrueStruct_exref = *(int *)_Py_TrueStruct_exref + 1;
        }
LAB_1802c74ab:
        local_res18 = local_res20;
        if ((-1 < *(int *)pcVar17) &&
           (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
          (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))();
        }
      }
    }
    local_res20 = plVar15;
    FUN_180382bf0(&local_res20);
    plVar15 = local_res20;
    lVar6 = *(longlong *)(param_1 + 0x10);
    if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar6 + 0x1a0) != 0)) &&
        (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) && (*(longlong *)(param_1 + 0x60) != 0)) {
LAB_1802c76f2:
      lVar6 = *(longlong *)(param_1 + 0x60);
      uVar16 = 0x73;
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar17 = pcVar11;
      goto LAB_1802c7969;
    }
    if (*(int *)(lVar6 + 0x174) != 0) {
      PyEval_SaveThread();
      PyEval_AcquireThread(param_1);
    }
    if (*(longlong *)(param_1 + 0x80) != 0) {
      plVar7 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x80) = 0;
      uVar16 = FUN_180373490(param_1);
      *(undefined8 *)(param_1 + 0x60) = uVar16;
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      goto LAB_1802c76f2;
    }
/* novels.services.spot_fix_patches#38="oi" */
    uVar10 = DAT_1803c4d60;
    pcVar17 = pcVar11;
    if (plVar15 == (longlong *)0x0) goto LAB_1802c7550;
    goto LAB_1802c71a0;
  }
LAB_1802c780c:
  lVar6 = FUN_180373860(plVar2);
  if (((-1 < lVar6) || (*(longlong *)(param_1 + 0x60) == 0)) &&
     (plVar7 = (longlong *)PyLong_FromSsize_t(lVar6), plVar7 != (longlong *)0x0)) {
    cVar4 = FUN_180384160(plVar15,plVar7);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    plVar14 = DAT_1803ab260;
    if (cVar4 != '\0') {
      plVar14 = plVar15;
    }
    iVar5 = (int)*plVar14 + 1;
    if (iVar5 != 0) {
      *(int *)plVar14 = iVar5;
    }
    FUN_180001d80(param_1);
    if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
       (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
      (**(code **)(local_res18[1] + 0x30))();
    }
    if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
       (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
      (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
    }
    if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
       (*local_60 = *local_60 + -1, *local_60 == 0)) {
      (**(code **)(local_60[1] + 0x30))();
    }
    if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
      (**(code **)(plVar2[1] + 0x30))(plVar2);
    }
    if ((int)*plVar1 < 0) {
      return plVar14;
    }
    *plVar1 = *plVar1 + -1;
    if (*plVar1 != 0) {
      return plVar14;
    }
    (**(code **)(plVar1[1] + 0x30))(plVar1);
    return plVar14;
  }
  uVar16 = 0x7f;
LAB_1802c795f:
  lVar6 = *(longlong *)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = 0;
LAB_1802c7969:
  piVar3 = *(int **)(lVar6 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar14,uVar16);
    plVar7 = *(longlong **)(lVar6 + 0x28);
    pcVar11 = pcVar9;
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1802c79cc:
      pcVar9 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar14) goto LAB_1802c7a15;
    pcVar9 = (code *)FUN_18038e2a0(plVar14,uVar16);
    *(int **)(pcVar9 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7 = *(longlong **)(lVar6 + 0x28);
    pcVar11 = pcVar9;
    if (pcVar9 != _Py_NoneStruct_exref) goto LAB_1802c79cc;
    pcVar11 = (code *)0x0;
  }
  *(code **)(lVar6 + 0x28) = pcVar11;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_1802c7a15:
  FUN_18036d0a0(plVar14,"oooooo",plVar2,plVar1,local_res18,pcVar17,plVar15,local_60);
  plVar7 = DAT_1803c4bb0;
  if (plVar14 == DAT_1803c4bb0) {
    if ((-1 < (int)*DAT_1803c4bb0) && (*DAT_1803c4bb0 = *DAT_1803c4bb0 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803c4bb0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
     (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
  if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
    (**(code **)(plVar15[1] + 0x30))(plVar15);
  }
  if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
     (*local_60 = *local_60 + -1, *local_60 == 0)) {
    (**(code **)(local_60[1] + 0x30))(local_60);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar6;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
