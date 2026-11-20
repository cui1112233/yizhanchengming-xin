/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800dd330 */

code * FUN_1800dd330(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  int iVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  undefined8 uVar11;
  code *pcVar12;
  longlong local_res18 [2];
  longlong *local_48;
  code *local_40;

  plVar3 = DAT_1803b3668;
  plVar1 = (longlong *)*param_3;
  pcVar12 = (code *)0x0;
  local_res18[0] = DAT_180391d08;
  if (DAT_1803b3668 == (longlong *)0x0) {
LAB_1800dd38e:
    DAT_1803b3668 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3660,DAT_1803ca790,0x10);
  }
  else if ((1 < *DAT_1803b3668) || (DAT_1803b3668[2] != 0)) {
    if ((DAT_1803b3668 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3668 && (*DAT_1803b3668 = *DAT_1803b3668 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1800dd38e;
  }
  plVar3 = DAT_1803b3668;
  FUN_180001d00(param_1,DAT_1803b3668);
  lVar5 = FUN_1800da7c0(param_1);
  if (lVar5 == 0) {
/* novels.pipeline.pipeline_checkpoints#3="FileRepository" */
    FUN_180373340(param_1,local_res18,DAT_1803b36c8);
    uVar11 = 0x2c;
    lVar5 = local_res18[0];
    pcVar12 = (code *)0x0;
  }
  else {
    plVar6 = (longlong *)FUN_18037f0c0();
    if (plVar6 == (longlong *)0x0) {
      lVar5 = *(longlong *)(param_1 + 0x60);
      uVar11 = 0x2c;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar5 = FUN_1800daa30(param_1);
      if (lVar5 == 0) {
/* novels.pipeline.pipeline_checkpoints#10="checkpoint_dir" */
        FUN_180373340(param_1,local_res18,DAT_1803b3700);
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
        uVar11 = 0x2c;
        lVar5 = local_res18[0];
      }
      else {
        uVar11 = 0x2c;
        *(undefined4 *)(plVar3 + 5) = 0x2c;
        plVar7 = (longlong *)FUN_180376770(param_1,lVar5,plVar1);
        if (plVar7 == (longlong *)0x0) {
          lVar5 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
        }
        else {
          lVar5 = FUN_1800da8d0(param_1);
          if (lVar5 == 0) {
/* novels.pipeline.pipeline_checkpoints#20="RUN_PAYLOAD_FILE" */
            FUN_180373340(param_1,local_res18,DAT_1803b3750);
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            lVar5 = local_res18[0];
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
              lVar5 = local_res18[0];
            }
          }
          else {
            plVar8 = (longlong *)FUN_180382540(plVar7,lVar5);
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if (plVar8 == (longlong *)0x0) {
              lVar5 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
            }
            else {
              local_40 = _Py_NoneStruct_exref;
              *(undefined4 *)(plVar3 + 5) = 0x2c;
              local_48 = plVar8;
              pcVar9 = (code *)FUN_180377750(param_1,plVar6,&local_48);
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if (pcVar9 == (code *)0x0) {
                lVar5 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
                iVar4 = PyObject_IsInstance(pcVar9,PyDict_Type_exref);
                if (iVar4 != -1) {
                  pcVar12 = _Py_NoneStruct_exref;
                  if (iVar4 != 0) {
                    pcVar12 = pcVar9;
                  }
                  if (*(int *)pcVar12 + 1 != 0) {
                    *(int *)pcVar12 = *(int *)pcVar12 + 1;
                  }
                  FUN_180001d80(param_1);
                  if ((-1 < *(int *)pcVar9) &&
                     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
                    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
                  }
                  if ((int)*plVar1 < 0) {
                    return pcVar12;
                  }
                  *plVar1 = *plVar1 + -1;
                  if (*plVar1 != 0) {
                    return pcVar12;
                  }
                  (**(code **)(plVar1[1] + 0x30))(plVar1);
                  return pcVar12;
                }
                lVar5 = *(longlong *)(param_1 + 0x60);
                uVar11 = 0x2d;
                *(undefined8 *)(param_1 + 0x60) = 0;
                pcVar12 = pcVar9;
              }
            }
          }
        }
      }
    }
  }
  piVar2 = *(int **)(lVar5 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar3,uVar11);
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar10 = pcVar9;
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
LAB_1800dd671:
      pcVar9 = pcVar10;
      if ((pcVar10 != (code *)0x0) && (*(int *)pcVar10 + 1 != 0)) {
        *(int *)pcVar10 = *(int *)pcVar10 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1800dd6ba;
    pcVar9 = (code *)FUN_18038e2a0(plVar3,uVar11);
    *(int **)(pcVar9 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar10 = pcVar9;
    if (pcVar9 != _Py_NoneStruct_exref) goto LAB_1800dd671;
    pcVar10 = (code *)0x0;
  }
  *(code **)(lVar5 + 0x28) = pcVar10;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_1800dd6ba:
  FUN_18036d0a0(plVar3,&DAT_180391be4,plVar1,pcVar12);
  plVar6 = DAT_1803b3668;
  if (plVar3 == DAT_1803b3668) {
    if ((-1 < (int)*DAT_1803b3668) && (*DAT_1803b3668 = *DAT_1803b3668 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b3668 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
