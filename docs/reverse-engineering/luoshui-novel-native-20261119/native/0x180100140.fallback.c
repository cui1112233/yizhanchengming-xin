/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180100140 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong * FUN_180100140(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  code *pcVar2;
  code *pcVar3;
  char cVar4;
  int iVar5;
  int iVar6;
  longlong *plVar7;
  code *pcVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  longlong *plVar12;
  longlong *plVar13;
  undefined8 uVar14;
  longlong lVar15;
  code *pcVar16;
  longlong *plVar17;
  undefined8 uVar18;
  undefined1 auStack_278 [32];
  longlong *local_258;
  longlong **local_250;
  longlong *local_248;
  code *local_240;
  code *local_238;
  code *local_230;
  longlong *local_228;
  longlong *local_220;
  longlong *local_218;
  code *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  code *local_1d0;
  code *local_1c8;
  longlong local_1b8;
  longlong *local_1b0;
  code *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  code *local_190;
  longlong *local_188;
  longlong *local_180;
  code *local_178;
  longlong *local_170;
  code *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  code *local_140;
  longlong *local_138;
  code *local_130;
  longlong *local_128;
  longlong *local_120;
  code *local_118;
  code *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  code *local_e8;
  longlong *local_e0;
  code *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  code *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  code *local_a0;
  longlong local_98;
  code *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong local_78;
  longlong *local_70;
  code *local_68;
  code *local_60;
  code *local_58;
  code *local_50;
  longlong *local_48;
  ulonglong local_40;
  
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_278;
  local_f0 = (longlong *)param_3[2];
  local_118 = (code *)param_3[5];
  plVar9 = (longlong *)*param_3;
  pcVar8 = (code *)param_3[1];
  local_108 = (longlong *)param_3[3];
  local_e0 = (longlong *)param_3[4];
  local_168 = (code *)0x0;
  local_110 = (code *)0x0;
  local_1a8 = pcVar8;
  local_198 = plVar9;
  local_1b0 = (longlong *)FUN_18036d6d0();
  plVar12 = DAT_1803b3b30;
  local_1b8 = DAT_180391d10;
  local_138 = (longlong *)0x0;
  local_150 = (longlong *)0x0;
  local_130 = (code *)0x0;
  local_100 = (longlong *)0x0;
  local_f8 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_a8 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_158 = (longlong *)0x0;
  local_148 = (longlong *)0x0;
  local_b8 = (code *)0x0;
  local_140 = (code *)0x0;
  local_170 = (longlong *)0x0;
  local_1a0 = DAT_1803b3b30;
  if (DAT_1803b3b30 == (longlong *)0x0) {
LAB_180100239:
    DAT_1803b3b30 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3d88,DAT_1803ca778,0xa8);
  }
  else if ((1 < *DAT_1803b3b30) || (DAT_1803b3b30[2] != 0)) {
    if ((DAT_1803b3b30 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3b30 && (*DAT_1803b3b30 = *DAT_1803b3b30 + -1, *plVar12 == 0)))) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    goto LAB_180100239;
  }
  plVar12 = DAT_1803b3b30;
  local_1a0 = DAT_1803b3b30;
  FUN_180001d00(param_1,DAT_1803b3b30);
  plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b3f10);
  if (plVar7 == (longlong *)0x0) {
    local_1b8 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar18 = 0x301;
    plVar7 = local_1a0;
    plVar12 = local_1b0;
  }
  else {
    uVar18 = 0x301;
    *(undefined4 *)(plVar12 + 5) = 0x301;
    local_d8 = (code *)FUN_18037e6c0(param_1,plVar7,DAT_1803b4268,pcVar8);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (local_d8 == (code *)0x0) {
      local_1b8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar7 = local_1a0;
      plVar12 = local_1b0;
    }
    else {
      local_168 = local_d8;
      plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b3f10);
      uVar18 = 0x302;
      if (plVar7 == (longlong *)0x0) {
        local_1b8 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar7 = local_1a0;
        plVar12 = local_1b0;
      }
      else {
        *(undefined4 *)(plVar12 + 5) = 0x302;
        pcVar8 = (code *)FUN_18037e6c0(param_1,plVar7,DAT_1803b4430,pcVar8);
        local_e8 = pcVar8;
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (pcVar8 == (code *)0x0) {
          local_1b8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar7 = local_1a0;
          pcVar8 = local_1a8;
          plVar12 = local_1b0;
        }
        else {
          local_110 = pcVar8;
          iVar5 = FUN_180001510(pcVar8);
          plVar7 = local_f0;
          uVar14 = DAT_1803b4af0;
          pcVar16 = PyExc_ValueError_exref;
          if (iVar5 == -1) {
            local_1b8 = *(longlong *)(param_1 + 0x60);
            uVar18 = 0x303;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar7 = local_1a0;
            pcVar8 = local_1a8;
            plVar12 = local_1b0;
          }
          else if (iVar5 == 0) {
            uVar18 = 0x304;
            *(undefined4 *)(plVar12 + 5) = 0x304;
            local_1b8 = FUN_180376770(param_1,*(undefined8 *)pcVar16,uVar14);
            FUN_18000d1d0(param_1,local_1b8);
            plVar7 = local_1a0;
            pcVar8 = local_1a8;
            plVar12 = local_1b0;
          }
          else {
            iVar5 = FUN_180001510(local_f0);
            if (iVar5 != -1) {
              if (iVar5 == 1) {
                iVar5 = (int)*plVar7 + 1;
                if (iVar5 != 0) {
                  *(int *)plVar7 = iVar5;
                }
LAB_18010045f:
                lVar11 = PyNumber_Long(plVar7);
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                if (lVar11 != 0) {
                  uVar18 = 0x306;
                  local_a0 = local_1a8;
                  local_1b0[2] = lVar11;
                  *(undefined4 *)(plVar12 + 5) = 0x306;
                  local_98 = lVar11;
                  local_120 = (longlong *)FUN_18037e820(param_1,plVar9,DAT_1803b4a68,&local_a0);
                  if (local_120 == (longlong *)0x0) {
                    local_1b8 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar7 = local_1a0;
                    pcVar8 = local_1a8;
                    plVar12 = local_1b0;
                  }
                  else {
                    local_138 = local_120;
                    plVar7 = (longlong *)FUN_18037f0c0(param_1,local_120,DAT_1803b3f38);
                    uVar18 = 0x307;
                    if (plVar7 == (longlong *)0x0) {
                      local_1b8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar7 = local_1a0;
                      pcVar8 = local_1a8;
                      plVar12 = local_1b0;
                    }
                    else {
                      *(undefined4 *)(plVar12 + 5) = 0x307;
                      plVar9 = (longlong *)FUN_180376210(param_1,plVar7);
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      if (plVar9 != (longlong *)0x0) {
                        iVar5 = FUN_180001510(plVar9);
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        piVar1 = DAT_1803b4a70;
                        if (iVar5 != -1) {
                          if (iVar5 == 0) {
                            plVar10 = (longlong *)FUN_180374a40(param_1,3);
                            if (*piVar1 + 1 != 0) {
                              *piVar1 = *piVar1 + 1;
                            }
                            plVar10[3] = (longlong)piVar1;
                            plVar7 = DAT_1803b3f30;
                            plVar9 = (longlong *)local_1b0[2];
                            if (DAT_1803ab740 == 0) {
                              DAT_1803ab740 = FUN_18035f440("format");
                            }
                            local_180 = plVar7;
                            local_188 = plVar9;
                            lVar11 = FUN_180377750(param_1,DAT_1803ab740,&local_188);
                            if (lVar11 == 0) {
                              uVar18 = 0x308;
LAB_1801005ef:
                              local_1b8 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar7 = local_1a0;
                              pcVar8 = local_1a8;
                              plVar9 = local_198;
                              plVar12 = local_1b0;
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                                plVar7 = local_1a0;
                                pcVar8 = local_1a8;
                                plVar9 = local_198;
                                plVar12 = local_1b0;
                              }
                            }
                            else {
                              plVar10[4] = lVar11;
                              piVar1 = DAT_1803b4a78;
                              if (*DAT_1803b4a78 + 1 != 0) {
                                *DAT_1803b4a78 = *DAT_1803b4a78 + 1;
                              }
                              plVar10[5] = (longlong)piVar1;
                              plVar9 = (longlong *)PyUnicode_Join(DAT_1803b3f30,plVar10);
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              pcVar8 = PyExc_ValueError_exref;
                              uVar18 = 0x308;
                              if (plVar9 == (longlong *)0x0) goto LAB_1801006c7;
                              *(undefined4 *)(plVar12 + 5) = 0x308;
                              lVar11 = FUN_180376770(param_1,*(undefined8 *)pcVar8,plVar9);
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              local_1b8 = lVar11;
                              FUN_18000d1d0(param_1,lVar11);
                              plVar7 = local_1a0;
                              pcVar8 = local_1a8;
                              plVar9 = local_198;
                              plVar12 = local_1b0;
                            }
                            goto LAB_1801006e3;
                          }
                          plVar9 = (longlong *)FUN_180001930(param_1,pcVar8);
                          if (plVar9 != (longlong *)0x0) {
                            local_d0 = (longlong *)FUN_18036d750(plVar9);
                            plVar7 = local_1b0;
                            local_c8 = local_1b0;
                            iVar5 = (int)*local_1b0 + 1;
                            if (iVar5 != 0) {
                              *(int *)local_1b0 = iVar5;
                            }
                            local_250 = &local_d0;
                            local_240 = (code *)0x48;
                            local_248 = (longlong *)0x2;
                            local_258 = (longlong *)DAT_1803b3d68;
                            plVar10 = (longlong *)
                                      FUN_180367890(FUN_1801037a0,DAT_1803ca778,DAT_1803b4a50,
                                                    DAT_1803b4b78);
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            local_128 = (longlong *)
                                        FUN_1800de860(param_1,plVar10,_Py_NoneStruct_exref);
                            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                            if (local_128 != (longlong *)0x0) {
                              local_150 = local_128;
                              iVar5 = FUN_180001510(local_128);
                              plVar9 = local_198;
                              piVar1 = DAT_1803b49b0;
                              if (iVar5 == -1) {
                                uVar18 = 0x30a;
                                goto LAB_1801006c7;
                              }
                              if (iVar5 == 0) {
                                plVar10 = (longlong *)FUN_180374a40(param_1,2);
                                if (*piVar1 + 1 != 0) {
                                  *piVar1 = *piVar1 + 1;
                                }
                                plVar10[3] = (longlong)piVar1;
                                plVar13 = DAT_1803b3f30;
                                plVar9 = (longlong *)plVar7[2];
                                if (DAT_1803ab740 == 0) {
                                  DAT_1803ab740 = FUN_18035f440("format");
                                }
                                local_180 = plVar13;
                                local_188 = plVar9;
                                lVar11 = FUN_180377750(param_1,DAT_1803ab740,&local_188);
                                if (lVar11 == 0) {
                                  uVar18 = 0x30b;
                                  goto LAB_1801005ef;
                                }
                                plVar10[4] = lVar11;
                                plVar9 = (longlong *)PyUnicode_Join(DAT_1803b3f30,plVar10);
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                pcVar8 = PyExc_ValueError_exref;
                                uVar18 = 0x30b;
                                if (plVar9 == (longlong *)0x0) goto LAB_1801006c7;
                                *(undefined4 *)(plVar12 + 5) = 0x30b;
                                lVar11 = FUN_180376770(param_1,*(undefined8 *)pcVar8,plVar9);
                                if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                }
                                local_1b8 = lVar11;
                                FUN_18000d1d0(param_1,lVar11);
                                plVar7 = local_1a0;
                                pcVar8 = local_1a8;
                                plVar9 = local_198;
                                plVar12 = local_1b0;
                              }
                              else {
                                plVar12 = (longlong *)FUN_18037f0c0(param_1,local_198,DAT_1803b3fa8)
                                ;
                                local_160 = plVar12;
                                if (plVar12 == (longlong *)0x0) {
                                  local_1b8 = *(longlong *)(param_1 + 0x60);
                                  uVar18 = 0x30d;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar7 = local_1a0;
                                  pcVar8 = local_1a8;
                                  plVar12 = local_1b0;
                                }
                                else {
                                  local_190 = DAT_1803b4718;
                                  local_178 = (code *)FUN_18037f0c0(param_1,local_d8,DAT_1803b4080);
                                  if (local_178 == (code *)0x0) {
                                    local_1b8 = *(longlong *)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    if ((-1 < (int)*plVar12) &&
                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                                    }
                                    uVar18 = 0x311;
                                    plVar7 = local_1a0;
                                    pcVar8 = local_1a8;
                                    plVar12 = local_1b0;
                                  }
                                  else {
                                    plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b4af8)
                                    ;
                                    if (plVar7 == (longlong *)0x0) {
                                      local_1b8 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      uVar18 = 0x312;
                                      plVar7 = local_1a0;
                                      pcVar8 = local_1a8;
                                      plVar12 = local_1b0;
                                    }
                                    else {
                                      plVar10 = (longlong *)FUN_180001930(param_1,plVar7);
                                      if ((-1 < (int)*plVar7) &&
                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                      }
                                      if (plVar10 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_1b8);
                                        uVar18 = 0x312;
                                        plVar7 = local_1a0;
                                        pcVar8 = local_1a8;
                                        plVar12 = local_1b0;
                                      }
                                      else {
                                        plVar13 = (longlong *)FUN_180372160(param_1,0);
                                        plVar9 = (longlong *)FUN_18000d430(plVar10);
                                        plVar7 = (longlong *)0x0;
                                        while (plVar9 != (longlong *)0x0) {
                                          plVar12 = plVar9;
                                          if (((local_170 != (longlong *)0x0) &&
                                              (-1 < (int)*local_170)) &&
                                             (*local_170 = *local_170 + -1, *local_170 == 0)) {
                                            plVar12 = local_170 + 1;
                                            local_170 = plVar9;
                                            (**(code **)(*plVar12 + 0x30))();
                                            plVar12 = local_170;
                                          }
                                          local_170 = plVar12;
                                          iVar5 = (int)*plVar9 + 1;
                                          if (iVar5 != 0) {
                                            *(int *)plVar9 = iVar5;
                                          }
                                          if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7))
                                             && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            (**(code **)(plVar7[1] + 0x30))();
                                          }
                                          uVar18 = DAT_1803b4090;
                                          uVar14 = _PyDict_NewPresized(2);
                                          PyDict_SetItem(uVar14,uVar18,plVar9);
                                          PyDict_SetItem(uVar14,DAT_1803b4098,DAT_1803b40a0);
                                          cVar4 = FUN_180372560(plVar13,uVar14);
                                          plVar7 = plVar9;
                                          if ((cVar4 == '\0') ||
                                             (cVar4 = FUN_18000d5b0(param_1), cVar4 == '\0'))
                                          goto LAB_180103309;
                                          plVar9 = (longlong *)FUN_18000d430(plVar10);
                                          plVar12 = local_160;
                                        }
                                        cVar4 = FUN_1800012a0(param_1);
                                        if (cVar4 == '\0') {
LAB_180103309:
                                          FUN_18000ce30(param_1,&local_1b8);
                                          lVar11 = local_1b8;
                                          FUN_18000ce70(&local_1b8);
                                          if ((-1 < (int)*plVar10) &&
                                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                                          }
                                          if ((-1 < (int)*plVar13) &&
                                             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                            (**(code **)(plVar13[1] + 0x30))(plVar13);
                                          }
                                          if ((plVar7 == (longlong *)0x0) || ((int)*plVar7 < 0)) {
                                            if (plVar7 != (longlong *)0x0) goto LAB_18010338b;
                                          }
                                          else {
                                            *plVar7 = *plVar7 + -1;
                                            if (*plVar7 == 0) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
LAB_18010338b:
                                            local_1b8 = 0;
                                            if ((-1 < (int)*plVar7) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
                                          }
                                          uVar18 = 0x312;
                                          plVar7 = local_1a0;
                                          pcVar8 = local_1a8;
                                          plVar9 = local_198;
                                          plVar12 = local_1b0;
                                          local_1b8 = lVar11;
                                        }
                                        else {
                                          iVar5 = (int)*plVar13 + 1;
                                          if (iVar5 != 0) {
                                            *(int *)plVar13 = iVar5;
                                          }
                                          if ((-1 < (int)*plVar10) &&
                                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                            lVar11 = FUN_18000cc40(plVar10);
                                            (**(code **)(lVar11 + 0x30))(plVar10);
                                          }
                                          if ((-1 < (int)*plVar13) &&
                                             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                            lVar11 = FUN_18000cc40(plVar13);
                                            (**(code **)(lVar11 + 0x30))(plVar13);
                                          }
                                          if ((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) {
                                            *plVar7 = *plVar7 + -1;
                                            iVar5 = (int)*plVar7;
                                            if (*plVar7 == 0) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              iVar5 = (int)*plVar7;
                                            }
                                            if ((-1 < iVar5) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
                                          }
                                          pcVar8 = local_178;
                                          uVar18 = 0x30d;
                                          local_50 = local_178;
                                          *(undefined4 *)(local_1a0 + 5) = 0x30d;
                                          local_60 = local_190;
                                          local_58 = local_190;
                                          local_258 = (longlong *)DAT_1803b4290;
                                          local_48 = plVar13;
                                          plVar12 = (longlong *)
                                                    FUN_18037d280(param_1,plVar12,DAT_1803b40a8,
                                                                  &local_60);
                                          plVar9 = local_160;
                                          if ((-1 < (int)*local_160) &&
                                             (*local_160 = *local_160 + -1, *local_160 == 0)) {
                                            lVar11 = FUN_18000cc40(local_160);
                                            (**(code **)(lVar11 + 0x30))(plVar9);
                                          }
                                          if ((-1 < *(int *)pcVar8) &&
                                             (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1,
                                             *(longlong *)pcVar8 == 0)) {
                                            lVar11 = FUN_18000cc40(pcVar8);
                                            (**(code **)(lVar11 + 0x30))(pcVar8);
                                          }
                                          if ((-1 < (int)*plVar13) &&
                                             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                            lVar11 = FUN_18000cc40(plVar13);
                                            (**(code **)(lVar11 + 0x30))(plVar13);
                                          }
                                          if (plVar12 == (longlong *)0x0) {
                                            FUN_18000ce30(param_1,&local_1b8);
                                            plVar7 = local_1a0;
                                            pcVar8 = local_1a8;
                                            plVar9 = local_198;
                                            plVar12 = local_1b0;
                                            goto LAB_1801006e3;
                                          }
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            lVar11 = FUN_18000cc40(plVar12);
                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                          }
                                          plVar9 = local_198;
                                          plVar10 = local_1a0;
                                          uVar18 = 0x314;
                                          *(undefined4 *)(local_1a0 + 5) = 0x314;
                                          plVar12 = (longlong *)
                                                    FUN_18037e980(param_1,local_198,DAT_1803b4060,
                                                                  DAT_1803b4b00 + 0x18);
                                          if (plVar12 == (longlong *)0x0) goto LAB_180100c7d;
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            lVar11 = FUN_18000cc40(plVar12);
                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                          }
                                          plVar12 = local_e0;
                                          iVar5 = FUN_180001510(local_e0);
                                          plVar7 = plVar10;
                                          if (iVar5 != -1) {
                                            plVar13 = DAT_1803b3f30;
                                            if (iVar5 == 1) {
                                              plVar13 = plVar12;
                                            }
                                            plVar12 = (longlong *)FUN_1803749a0(plVar13);
                                            if (plVar12 != (longlong *)0x0) {
                                              plVar13 = (longlong *)
                                                        FUN_18037f0c0(param_1,plVar12,DAT_1803b3f38)
                                              ;
                                              if ((-1 < (int)*plVar12) &&
                                                 (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar12);
                                                (**(code **)(lVar11 + 0x30))(plVar12);
                                              }
                                              if (plVar13 != (longlong *)0x0) {
                                                uVar18 = 0x315;
                                                *(undefined4 *)(plVar10 + 5) = 0x315;
                                                pcVar8 = (code *)FUN_180376210(param_1,plVar13);
                                                local_190 = pcVar8;
                                                if ((-1 < (int)*plVar13) &&
                                                   (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                  lVar11 = FUN_18000cc40(plVar13);
                                                  (**(code **)(lVar11 + 0x30))(plVar13);
                                                }
                                                if (pcVar8 == (code *)0x0) {
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                }
                                                else {
                                                  local_130 = pcVar8;
                                                  iVar5 = FUN_180001510(pcVar8);
                                                  plVar9 = local_198;
                                                  if (iVar5 == -1) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x316;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else if (iVar5 == 0) {
                                                    uVar18 = 0x317;
                                                    *(undefined4 *)(plVar10 + 5) = 0x317;
                                                    plVar12 = (longlong *)
                                                              FUN_18037e6c0(param_1,local_198,
                                                                            DAT_1803b4260,
                                                                            *(undefined8 *)
                                                                             (DAT_1803b4b08 + 0x18))
                                                    ;
                                                    if (plVar12 == (longlong *)0x0)
                                                    goto LAB_180100c7d;
                                                    if ((-1 < (int)*plVar12) &&
                                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar12);
                                                      (**(code **)(lVar11 + 0x30))(plVar12);
                                                    }
                                                    lVar11 = FUN_1800dec70(param_1);
                                                    if (lVar11 == 0) {
                                                      FUN_180373340(param_1,&local_1b8,DAT_1803b43f8
                                                                   );
                                                      uVar18 = 0x318;
                                                      pcVar8 = local_1a8;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_18037f0c0(param_1,plVar9,
                                                                             DAT_1803b3fd8);
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x318;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        plVar13 = (longlong *)
                                                                  FUN_180374a40(param_1,2);
                                                        if (*(int *)local_1a8 + 1 != 0) {
                                                          *(int *)local_1a8 = *(int *)local_1a8 + 1;
                                                        }
                                                        FUN_18000cc80(plVar13,0,local_1a8);
                                                        pcVar8 = local_118;
                                                        iVar5 = FUN_180001510(local_118);
                                                        if (iVar5 == -1) {
LAB_180100fa6:
                                                          uVar18 = 0x318;
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar11 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar11 + 0x30))(plVar9);
                                                          }
                                                          plVar7 = local_1a0;
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                          if ((-1 < (int)*plVar13) &&
                                                             (*plVar13 = *plVar13 + -1,
                                                             *plVar13 == 0)) {
                                                            lVar11 = FUN_18000cc40(plVar13);
                                                            (**(code **)(lVar11 + 0x30))(plVar13);
                                                            plVar7 = plVar10;
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                        }
                                                        else {
                                                          if (iVar5 == 1) {
                                                            if (*(int *)pcVar8 + 1 != 0) {
                                                              *(int *)pcVar8 = *(int *)pcVar8 + 1;
                                                            }
                                                          }
                                                          else {
                                                            pcVar8 = (code *)FUN_18037f0c0(param_1,
                                                  local_168,DAT_1803b42a0);
                                                  if (pcVar8 == (code *)0x0) goto LAB_180100fa6;
                                                  }
                                                  FUN_18000cc80(plVar13,1,pcVar8);
                                                  plVar12 = (longlong *)
                                                            FUN_18036fb80(param_1,DAT_1803b4400);
                                                  uVar18 = 0x318;
                                                  *(undefined4 *)(plVar10 + 5) = 0x318;
                                                  plVar10 = (longlong *)
                                                            FUN_180001790(param_1,plVar9,plVar13,
                                                                          plVar12);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar15 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar13) &&
                                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar13);
                                                    (**(code **)(lVar15 + 0x30))(plVar13);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar15 + 0x30))(plVar12);
                                                  }
                                                  plVar7 = local_1a0;
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    *(undefined4 *)(local_1a0 + 5) = 0x318;
                                                    plVar9 = (longlong *)
                                                             FUN_180376770(param_1,lVar11,plVar10);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                    }
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      local_100 = plVar9;
                                                      pcVar16 = (code *)FUN_18037f0c0(param_1,plVar9
                                                                                      ,DAT_1803b4528
                                                                                     );
                                                      plVar12 = local_1b0;
                                                      if (pcVar16 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x319;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        lVar11 = local_1b0[2];
                                                        plVar9 = (longlong *)
                                                                 FUN_18037f0c0(param_1,local_198,
                                                                               DAT_1803b3f10);
                                                        plVar10 = local_1a0;
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          if ((-1 < *(int *)pcVar16) &&
                                                             (*(longlong *)pcVar16 =
                                                                   *(longlong *)pcVar16 + -1,
                                                             *(longlong *)pcVar16 == 0)) {
                                                            lVar11 = FUN_18000cc40(pcVar16);
                                                            (**(code **)(lVar11 + 0x30))(pcVar16);
                                                          }
                                                          uVar18 = 0x319;
                                                          plVar7 = local_1a0;
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                        }
                                                        else {
                                                          uVar18 = 0x319;
                                                          *(undefined4 *)(local_1a0 + 5) = 0x319;
                                                          plVar12 = (longlong *)
                                                                    FUN_18037e6c0(param_1,plVar9,
                                                                                  DAT_1803b41c8,
                                                                                  local_1a8);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar15 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar15 + 0x30))(plVar9);
                                                          }
                                                          plVar7 = plVar10;
                                                          if (plVar12 == (longlong *)0x0)
                                                          goto LAB_180101139;
                                                          local_80 = local_138;
                                                          local_90 = local_168;
                                                          *(undefined4 *)(plVar10 + 5) = 0x319;
                                                          local_258 = (longlong *)DAT_1803b46b0;
                                                          local_88 = plVar12;
                                                          local_78 = lVar11;
                                                          plVar13 = (longlong *)
                                                                    FUN_18037d650(param_1,pcVar16,
                                                                                  &local_80,
                                                                                  &local_90);
                                                          if ((-1 < *(int *)pcVar16) &&
                                                             (*(longlong *)pcVar16 =
                                                                   *(longlong *)pcVar16 + -1,
                                                             *(longlong *)pcVar16 == 0)) {
                                                            lVar11 = FUN_18000cc40(pcVar16);
                                                            (**(code **)(lVar11 + 0x30))(pcVar16);
                                                          }
                                                          if ((-1 < (int)*plVar12) &&
                                                             (*plVar12 = *plVar12 + -1,
                                                             *plVar12 == 0)) {
                                                            lVar11 = FUN_18000cc40(plVar12);
                                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                                          }
                                                          plVar9 = local_198;
                                                          if (plVar13 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                          else {
                                                            uVar18 = 0x31a;
                                                            local_c8 = (longlong *)local_1b0[2];
                                                            *(undefined4 *)(plVar10 + 5) = 0x31a;
                                                            local_f8 = plVar13;
                                                            local_d0 = plVar13;
                                                            plVar12 = (longlong *)
                                                                      FUN_18037e820(param_1,
                                                  local_198,DAT_1803b4530,&local_d0);
                                                  if (plVar12 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar12) &&
                                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar12);
                                                      (**(code **)(lVar11 + 0x30))(plVar12);
                                                    }
                                                    plVar9 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar13,
                                                                           DAT_1803b3ff0);
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      uVar18 = 0x31d;
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      uVar18 = 0x31d;
                                                      *(undefined4 *)(plVar10 + 5) = 0x31d;
                                                      pcVar16 = (code *)FUN_1803770d0(param_1,plVar9
                                                                                      ,DAT_1803b45a0
                                                                                     );
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar11 + 0x30))(plVar9);
                                                      }
                                                      if (pcVar16 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        iVar5 = FUN_180001510(pcVar16);
                                                        if (iVar5 == -1) {
LAB_180101139:
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                          if ((-1 < *(int *)pcVar16) &&
                                                             (*(longlong *)pcVar16 =
                                                                   *(longlong *)pcVar16 + -1,
                                                             *(longlong *)pcVar16 == 0)) {
                                                            lVar11 = FUN_18000cc40(pcVar16);
                                                            (**(code **)(lVar11 + 0x30))(pcVar16);
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                        }
                                                        else {
                                                          if ((-1 < *(int *)pcVar16) &&
                                                             (*(longlong *)pcVar16 =
                                                                   *(longlong *)pcVar16 + -1,
                                                             *(longlong *)pcVar16 == 0)) {
                                                            lVar11 = FUN_18000cc40(pcVar16);
                                                            (**(code **)(lVar11 + 0x30))(pcVar16);
                                                          }
                                                          plVar9 = (longlong *)
                                                                   FUN_18037f0c0(param_1,plVar13,
                                                                                 DAT_1803b3ff0);
                                                          if (iVar5 == 0) {
                                                            if (plVar9 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              uVar18 = 0x31e;
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                            }
                                                            else {
                                                              uVar18 = 0x31e;
                                                              *(undefined4 *)(plVar10 + 5) = 0x31e;
                                                              pcVar16 = (code *)FUN_1803770d0(
                                                  param_1,plVar9,DAT_1803b4b10);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar11 + 0x30))(plVar9);
                                                  }
                                                  if (pcVar16 != (code *)0x0) {
                                                    iVar5 = FUN_180001510(pcVar16);
                                                    if (iVar5 != -1) {
                                                      if (iVar5 != 1) {
                                                        if ((-1 < *(int *)pcVar16) &&
                                                           (*(longlong *)pcVar16 =
                                                                 *(longlong *)pcVar16 + -1,
                                                           *(longlong *)pcVar16 == 0)) {
                                                          lVar11 = FUN_18000cc40(pcVar16);
                                                          (**(code **)(lVar11 + 0x30))(pcVar16);
                                                        }
                                                        pcVar16 = DAT_1803b4b18;
                                                        if (*(int *)DAT_1803b4b18 + 1 != 0) {
                                                          *(int *)DAT_1803b4b18 =
                                                               *(int *)DAT_1803b4b18 + 1;
                                                        }
                                                      }
                                                      goto LAB_1801014b3;
                                                    }
                                                    goto LAB_180101139;
                                                  }
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  else if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x31c;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    uVar18 = 0x31c;
                                                    *(undefined4 *)(plVar10 + 5) = 0x31c;
                                                    pcVar16 = (code *)FUN_1803770d0(param_1,plVar9,
                                                                                    DAT_1803b4b10);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar11 + 0x30))(plVar9);
                                                    }
                                                    if (pcVar16 != (code *)0x0) {
LAB_1801014b3:
                                                      pcVar8 = local_190;
                                                      local_130 = pcVar16;
                                                      if ((-1 < *(int *)local_190) &&
                                                         (*(longlong *)local_190 =
                                                               *(longlong *)local_190 + -1,
                                                         *(longlong *)local_190 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_190);
                                                        (**(code **)(lVar11 + 0x30))(pcVar8);
                                                      }
                                                      goto LAB_1801014d7;
                                                    }
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
LAB_1801014d7:
                                                    plVar9 = local_198;
                                                    uVar18 = 800;
                                                    *(undefined4 *)(plVar10 + 5) = 800;
                                                    plVar12 = (longlong *)
                                                              FUN_18037e820(param_1,local_198,
                                                                            DAT_1803b4060,
                                                                            DAT_1803b40b8 + 0x18);
                                                    if (plVar12 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      plVar7 = local_1a0;
                                                      pcVar8 = local_1a8;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar12) &&
                                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar12);
                                                        (**(code **)(lVar11 + 0x30))(plVar12);
                                                      }
                                                      uVar18 = 0x322;
                                                      *(undefined4 *)(plVar10 + 5) = 0x322;
                                                      plVar12 = (longlong *)
                                                                FUN_18037e980(param_1,plVar9,
                                                                              DAT_1803b4060,
                                                                              DAT_1803b4b20 + 0x18);
                                                      if (plVar12 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        plVar7 = local_1a0;
                                                        pcVar8 = local_1a8;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar11 + 0x30))(plVar12);
                                                        }
                                                        lVar11 = FUN_1800df820(param_1);
                                                        if (lVar11 == 0) {
                                                          FUN_180373340(param_1,&local_1b8,
                                                                        DAT_1803b4408);
                                                          uVar18 = 0x323;
                                                          plVar7 = local_1a0;
                                                          pcVar8 = local_1a8;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          plVar9 = (longlong *)
                                                                   FUN_18037f0c0(param_1,plVar9,
                                                                                 DAT_1803b3fd8);
                                                          if (plVar9 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            uVar18 = 0x323;
                                                            plVar7 = local_1a0;
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_180374a40(param_1,2);
                                                            if (*(int *)local_1a8 + 1 != 0) {
                                                              *(int *)local_1a8 =
                                                                   *(int *)local_1a8 + 1;
                                                            }
                                                            FUN_18000cc80(plVar13,0,local_1a8);
                                                            pcVar8 = local_118;
                                                            iVar5 = FUN_180001510(local_118);
                                                            if (iVar5 == -1) {
LAB_180101711:
                                                              uVar18 = 0x323;
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              if ((-1 < (int)*plVar9) &&
                                                                 (*plVar9 = *plVar9 + -1,
                                                                 *plVar9 == 0)) {
                                                                lVar11 = FUN_18000cc40(plVar9);
                                                                (**(code **)(lVar11 + 0x30))(plVar9)
                                                                ;
                                                              }
LAB_18010173f:
                                                              plVar7 = local_1a0;
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
LAB_180101752:
                                                                lVar11 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar11 + 0x30))
                                                                          (plVar13);
                                                                plVar7 = local_1a0;
                                                                pcVar8 = local_1a8;
                                                                plVar9 = local_198;
                                                                plVar12 = local_1b0;
                                                              }
                                                            }
                                                            else {
                                                              if (iVar5 == 1) {
                                                                if (*(int *)pcVar8 + 1 != 0) {
                                                                  *(int *)pcVar8 =
                                                                       *(int *)pcVar8 + 1;
                                                                }
                                                              }
                                                              else {
                                                                pcVar8 = (code *)FUN_18037f0c0(
                                                  param_1,local_168,DAT_1803b42a0);
                                                  if (pcVar8 == (code *)0x0) goto LAB_180101711;
                                                  }
                                                  FUN_18000cc80(plVar13,1,pcVar8);
                                                  plVar12 = (longlong *)
                                                            FUN_18036fb80(param_1,DAT_1803b4410);
                                                  uVar18 = 0x323;
                                                  *(undefined4 *)(plVar10 + 5) = 0x323;
                                                  plVar10 = (longlong *)
                                                            FUN_180001790(param_1,plVar9,plVar13,
                                                                          plVar12);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar15 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar13) &&
                                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar13);
                                                    (**(code **)(lVar15 + 0x30))(plVar13);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar15 + 0x30))(plVar12);
                                                  }
                                                  plVar7 = local_1a0;
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    *(undefined4 *)(local_1a0 + 5) = 0x323;
                                                    plVar9 = (longlong *)
                                                             FUN_180376770(param_1,lVar11,plVar10);
                                                    local_d0 = plVar9;
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                    }
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      local_b0 = plVar9;
                                                      plVar12 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar9,
                                                                              DAT_1803b4720);
                                                      plVar9 = local_108;
                                                      if (plVar12 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x324;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        iVar5 = FUN_180001510(local_108);
                                                        if (iVar5 != -1) {
                                                          plVar10 = DAT_1803b4728;
                                                          if (iVar5 == 1) {
                                                            plVar10 = plVar9;
                                                          }
                                                          pcVar8 = (code *)FUN_1803749a0(plVar10);
                                                          if (pcVar8 != (code *)0x0) {
                                                            local_70 = local_138;
                                                            uVar18 = 0x324;
                                                            local_68 = local_130;
                                                            *(undefined4 *)(plVar7 + 5) = 0x324;
                                                            local_258 = (longlong *)DAT_1803b4b28;
                                                            local_178 = pcVar8;
                                                            plVar10 = (longlong *)
                                                                      FUN_18037d650(param_1,plVar12,
                                                                                    &local_70,
                                                                                    &local_178);
                                                            local_188 = plVar10;
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            if ((-1 < *(int *)pcVar8) &&
                                                               (*(longlong *)pcVar8 =
                                                                     *(longlong *)pcVar8 + -1,
                                                               *(longlong *)pcVar8 == 0)) {
                                                              lVar11 = FUN_18000cc40(pcVar8);
                                                              (**(code **)(lVar11 + 0x30))(pcVar8);
                                                            }
                                                            plVar9 = local_198;
                                                            if (plVar10 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                            }
                                                            else {
                                                              uVar18 = 0x325;
                                                              *(undefined4 *)(plVar7 + 5) = 0x325;
                                                              local_a8 = plVar10;
                                                              plVar12 = (longlong *)
                                                                        FUN_18037e820(param_1,
                                                  local_198,DAT_1803b4060,DAT_1803b4170 + 0x18);
                                                  if (plVar12 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar12) &&
                                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar12);
                                                      (**(code **)(lVar11 + 0x30))(plVar12);
                                                    }
                                                    uVar18 = 0x327;
                                                    *(undefined4 *)(plVar7 + 5) = 0x327;
                                                    plVar12 = (longlong *)
                                                              FUN_18037e980(param_1,plVar9,
                                                                            DAT_1803b4060,
                                                                            DAT_1803b4b30 + 0x18);
                                                    if (plVar12 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar12) &&
                                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar12);
                                                        (**(code **)(lVar11 + 0x30))(plVar12);
                                                      }
                                                      plVar12 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar10,
                                                                              DAT_1803b3f38);
                                                      if (plVar12 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x328;
                                                        pcVar8 = local_1a8;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        uVar18 = 0x328;
                                                        *(undefined4 *)(plVar7 + 5) = 0x328;
                                                        plVar9 = (longlong *)
                                                                 FUN_180376210(param_1,plVar12);
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar11 + 0x30))(plVar12);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          plVar12 = (longlong *)
                                                                    FUN_18037f0c0(param_1,plVar9,
                                                                                  DAT_1803b49b8);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar11 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar11 + 0x30))(plVar9);
                                                          }
                                                          if (plVar12 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                          else {
                                                            *(undefined4 *)(plVar7 + 5) = 0x328;
                                                            plVar9 = (longlong *)
                                                                     FUN_180376210(param_1,plVar12);
                                                            local_170 = plVar9;
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            if (plVar9 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                            }
                                                            else {
                                                              local_c0 = plVar9;
                                                              plVar12 = (longlong *)
                                                                        FUN_18037f0c0(param_1,
                                                  local_150,DAT_1803b4080);
                                                  if (plVar12 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x329;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    local_158 = plVar12;
                                                    plVar10 = (longlong *)
                                                              FUN_18037f0c0(param_1,plVar10,
                                                                            DAT_1803b3f38);
                                                    plVar7 = local_1a0;
                                                    if (plVar10 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      uVar18 = 0x32a;
                                                      plVar7 = local_1a0;
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      uVar18 = 0x32a;
                                                      *(undefined4 *)(local_1a0 + 5) = 0x32a;
                                                      plVar13 = (longlong *)
                                                                FUN_180376210(param_1,plVar10);
                                                      if ((-1 < (int)*plVar10) &&
                                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar10);
                                                        (**(code **)(lVar11 + 0x30))(plVar10);
                                                      }
                                                      if (plVar13 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        local_148 = plVar13;
                                                        iVar5 = FUN_180001510(plVar9);
                                                        if (iVar5 == -1) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          uVar18 = 0x32b;
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          iVar6 = 0;
                                                          if (iVar5 == 0) {
LAB_180101d24:
                                                            plVar10 = plVar13;
                                                            if (iVar6 != 0) {
                                                              plVar9 = (longlong *)
                                                                       FUN_1800b4830(param_1,plVar9,
                                                                                     DAT_1803ab250,0
                                                                                    );
                                                              if (plVar9 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_1b8);
                                                                uVar18 = 0x32c;
                                                                pcVar8 = local_1a8;
                                                                plVar9 = local_198;
                                                                plVar12 = local_1b0;
                                                              }
                                                              else {
                                                                plVar10 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
                                                  plVar9,DAT_1803b49d0);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar11 + 0x30))(plVar9);
                                                  }
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x32c;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    uVar18 = 0x32c;
                                                    *(undefined4 *)(plVar7 + 5) = 0x32c;
                                                    plVar9 = (longlong *)
                                                             FUN_1803770d0(param_1,plVar10,
                                                                           DAT_1803b49c8);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                    }
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      plVar7 = local_1a0;
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      plVar7 = (longlong *)
                                                               FUN_18037f0c0(param_1,plVar9,
                                                                             DAT_1803b3f38);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar11 + 0x30))(plVar9);
                                                      }
                                                      if (plVar7 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        plVar7 = local_1a0;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        *(undefined4 *)(local_1a0 + 5) = 0x32c;
                                                        plVar10 = (longlong *)
                                                                  FUN_180376210(param_1,plVar7);
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          lVar11 = FUN_18000cc40(plVar7);
                                                          (**(code **)(lVar11 + 0x30))(plVar7);
                                                        }
                                                        if (plVar10 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          plVar7 = local_1a0;
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          iVar5 = FUN_180001510(plVar10);
                                                          if (iVar5 == -1) {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            plVar7 = local_1a0;
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                            if ((-1 < (int)*plVar10) &&
                                                               (*plVar10 = *plVar10 + -1,
                                                               *plVar10 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar10);
                                                              (**(code **)(lVar11 + 0x30))(plVar10);
                                                              plVar7 = local_1a0;
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                            }
                                                          }
                                                          else {
                                                            if (iVar5 != 1) {
                                                              if ((-1 < (int)*plVar10) &&
                                                                 (*plVar10 = *plVar10 + -1,
                                                                 *plVar10 == 0)) {
                                                                lVar11 = FUN_18000cc40(plVar10);
                                                                (**(code **)(lVar11 + 0x30))
                                                                          (plVar10);
                                                              }
                                                              iVar5 = (int)*plVar12 + 1;
                                                              plVar10 = plVar12;
                                                              if (iVar5 != 0) {
                                                                *(int *)plVar12 = iVar5;
                                                              }
                                                            }
                                                            local_158 = plVar10;
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            uVar18 = DAT_1803b4398;
                                                            plVar9 = (longlong *)
                                                                     FUN_180024e50(param_1,local_170
                                                                                   ,DAT_1803b49d8);
                                                            if (plVar9 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              uVar18 = 0x32d;
                                                              plVar7 = local_1a0;
                                                              pcVar8 = local_1a8;
                                                              plVar9 = local_198;
                                                              plVar12 = local_1b0;
                                                            }
                                                            else {
                                                              plVar12 = (longlong *)
                                                                        FUN_180374990(param_1,uVar18
                                                                                      ,plVar9);
                                                              if ((-1 < (int)*plVar9) &&
                                                                 (*plVar9 = *plVar9 + -1,
                                                                 *plVar9 == 0)) {
                                                                lVar11 = FUN_18000cc40(plVar9);
                                                                (**(code **)(lVar11 + 0x30))(plVar9)
                                                                ;
                                                              }
                                                              if (plVar12 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_1b8);
                                                                uVar18 = 0x32d;
                                                                plVar7 = local_1a0;
                                                                pcVar8 = local_1a8;
                                                                plVar9 = local_198;
                                                                plVar12 = local_1b0;
                                                              }
                                                              else {
                                                                plVar10 = (longlong *)
                                                                          FUN_18036dfb0(param_1,
                                                  plVar12);
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  if (plVar10 != (longlong *)0x0) {
                                                    local_148 = plVar10;
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar11 + 0x30))(plVar13);
                                                    }
                                                    goto LAB_180102019;
                                                  }
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  uVar18 = 0x32d;
                                                  plVar7 = local_1a0;
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801006e3;
                                                  }
LAB_180102019:
                                                  piVar1 = DAT_1803b4350;
                                                  plVar13 = (longlong *)FUN_180374a40(param_1,4);
                                                  if (*piVar1 + 1 != 0) {
                                                    *piVar1 = *piVar1 + 1;
                                                  }
                                                  FUN_18000cc80(plVar13,0,piVar1);
                                                  lVar11 = FUN_180375520(param_1,local_158,
                                                                         DAT_1803b3f30);
                                                  if (lVar11 == 0) {
LAB_1801020a4:
                                                    uVar18 = 0x32e;
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    goto LAB_18010173f;
                                                  }
                                                  FUN_18000cc80(plVar13,1,lVar11);
                                                  if (*DAT_1803b4358 + 1 != 0) {
                                                    *DAT_1803b4358 = *DAT_1803b4358 + 1;
                                                  }
                                                  FUN_18000cc80(plVar13,2);
                                                  lVar11 = FUN_180375520(param_1,plVar10,
                                                                         DAT_1803b3f30);
                                                  if (lVar11 == 0) goto LAB_1801020a4;
                                                  FUN_18000cc80(plVar13,3,lVar11);
                                                  plVar9 = (longlong *)
                                                           PyUnicode_Join(DAT_1803b3f30,plVar13);
                                                  if ((-1 < (int)*plVar13) &&
                                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar13);
                                                    (**(code **)(lVar11 + 0x30))(plVar13);
                                                  }
                                                  if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x32e;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    goto LAB_1801006e3;
                                                  }
                                                  pcVar8 = (code *)FUN_18036dfb0(param_1,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar11 + 0x30))(plVar9);
                                                  }
                                                  if (pcVar8 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x32e;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    goto LAB_1801006e3;
                                                  }
                                                  local_b8 = pcVar8;
                                                  lVar11 = FUN_1800df1c0(param_1);
                                                  if (lVar11 == 0) {
                                                    FUN_180373340(param_1,&local_1b8,DAT_1803b41b8);
                                                    uVar18 = 0x32f;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    goto LAB_1801006e3;
                                                  }
                                                  plVar13 = (longlong *)
                                                            FUN_18037f0c0(param_1,lVar11,
                                                                          DAT_1803b41c0);
                                                  plVar9 = local_198;
                                                  if (plVar13 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x32f;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    goto LAB_1801006e3;
                                                  }
                                                  plVar12 = (longlong *)
                                                            FUN_18037f0c0(param_1,local_198,
                                                                          DAT_1803b3f10);
                                                  plVar7 = local_1a0;
                                                  if (plVar12 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar11 + 0x30))(plVar13);
                                                    }
                                                    uVar18 = 0x32f;
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar12 = local_1b0;
                                                    goto LAB_1801006e3;
                                                  }
                                                  uVar18 = 0x32f;
                                                  *(undefined4 *)(local_1a0 + 5) = 0x32f;
                                                  plVar10 = (longlong *)
                                                            FUN_18037e6c0(param_1,plVar12,
                                                                          DAT_1803b4438,local_1a8);
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar11 + 0x30))(plVar13);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    goto LAB_1801006e3;
                                                  }
                                                  plVar9 = (longlong *)
                                                           FUN_18037f0c0(param_1,local_150,
                                                                         DAT_1803b4450);
                                                  if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar11 + 0x30))(plVar13);
                                                    }
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    goto LAB_1801006e3;
                                                  }
                                                  pcVar16 = (code *)FUN_180382540(plVar10,plVar9);
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar10);
                                                    (**(code **)(lVar11 + 0x30))(plVar10);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar11 + 0x30))(plVar9);
                                                  }
                                                  plVar10 = local_1a0;
                                                  if (pcVar16 != (code *)0x0) {
                                                    *(undefined4 *)(local_1a0 + 5) = 0x32f;
                                                    local_60 = pcVar16;
                                                    local_58 = pcVar8;
                                                    plVar9 = (longlong *)
                                                             FUN_180377750(param_1,plVar13,&local_60
                                                                          );
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar11 + 0x30))(plVar13);
                                                    }
                                                    if ((-1 < *(int *)pcVar16) &&
                                                       (*(longlong *)pcVar16 =
                                                             *(longlong *)pcVar16 + -1,
                                                       *(longlong *)pcVar16 == 0)) {
                                                      lVar11 = FUN_18000cc40(pcVar16);
                                                      (**(code **)(lVar11 + 0x30))(pcVar16);
                                                    }
                                                    plVar7 = plVar10;
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar11 + 0x30))(plVar9);
                                                      }
                                                      plVar9 = local_150;
                                                      cVar4 = FUN_18037f3a0(param_1,local_150,
                                                                            DAT_1803b4080,local_158)
                                                      ;
                                                      if (cVar4 == '\0') {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x330;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        plVar12 = (longlong *)
                                                                  FUN_180375370(param_1,local_148);
                                                        if (plVar12 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          uVar18 = 0x331;
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          cVar4 = FUN_18037f3a0(param_1,plVar9,
                                                                                DAT_1803b45f8,
                                                                                plVar12);
                                                          if ((-1 < (int)*plVar12) &&
                                                             (*plVar12 = *plVar12 + -1,
                                                             *plVar12 == 0)) {
                                                            lVar11 = FUN_18000cc40(plVar12);
                                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                                          }
                                                          plVar9 = local_198;
                                                          if (cVar4 == '\0') {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            uVar18 = 0x331;
                                                            pcVar8 = local_1a8;
                                                            plVar9 = local_198;
                                                            plVar12 = local_1b0;
                                                          }
                                                          else {
                                                            uVar18 = 0x332;
                                                            *(undefined4 *)(plVar10 + 5) = 0x332;
                                                            plVar12 = (longlong *)
                                                                      FUN_18037e820(param_1,
                                                  local_198,DAT_1803b4060,DAT_1803b4178 + 0x18);
                                                  if (plVar12 == (longlong *)0x0)
                                                  goto LAB_180100c7d;
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  pcVar16 = local_118;
                                                  iVar5 = FUN_180001510(local_118);
                                                  if (iVar5 == -1) {
LAB_18010252d:
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x334;
                                                    pcVar8 = local_1a8;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    if (iVar5 == 1) {
                                                      if (*(int *)pcVar16 + 1 != 0) {
                                                        *(int *)pcVar16 = *(int *)pcVar16 + 1;
                                                      }
                                                      local_140 = pcVar16;
                                                      pcVar16 = local_140;
                                                    }
                                                    else {
                                                      pcVar16 = (code *)FUN_18037f0c0(param_1,
                                                  local_168,DAT_1803b42a0);
                                                  if (pcVar16 == (code *)0x0) goto LAB_18010252d;
                                                  }
                                                  local_140 = pcVar16;
                                                  uVar18 = 0x335;
                                                  *(undefined4 *)(plVar10 + 5) = 0x335;
                                                  plVar12 = (longlong *)
                                                            FUN_18037e980(param_1,plVar9,
                                                                          DAT_1803b4060,
                                                                          DAT_1803b4b38 + 0x18);
                                                  if (plVar12 == (longlong *)0x0)
                                                  goto LAB_180100c7d;
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  plVar10 = (longlong *)
                                                            FUN_18037f0c0(param_1,plVar9,
                                                                          DAT_1803b4260);
                                                  piVar1 = DAT_1803b4b40;
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    uVar18 = 0x336;
                                                    pcVar8 = local_1a8;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    plVar13 = (longlong *)FUN_180374a40(param_1,3);
                                                    if (*piVar1 + 1 != 0) {
                                                      *piVar1 = *piVar1 + 1;
                                                    }
                                                    FUN_18000cc80(plVar13,0,piVar1);
                                                    plVar12 = local_1b0;
                                                    lVar11 = FUN_180375520(param_1,local_1b0[2],
                                                                           DAT_1803b3f30);
                                                    if (lVar11 == 0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      uVar18 = 0x336;
                                                      if ((-1 < (int)*plVar10) &&
                                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar10);
                                                        (**(code **)(lVar11 + 0x30))(plVar10);
                                                      }
                                                      plVar7 = local_1a0;
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      if ((-1 < (int)*plVar13) &&
                                                         (*plVar13 = *plVar13 + -1, *plVar13 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar13);
                                                        (**(code **)(lVar11 + 0x30))(plVar13);
                                                        plVar7 = local_1a0;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                      }
                                                    }
                                                    else {
                                                      FUN_18000cc80(plVar13,1,lVar11);
                                                      if (*DAT_1803b4b48 + 1 != 0) {
                                                        *DAT_1803b4b48 = *DAT_1803b4b48 + 1;
                                                      }
                                                      FUN_18000cc80(plVar13,2);
                                                      plVar9 = (longlong *)
                                                               PyUnicode_Join(DAT_1803b3f30,plVar13)
                                                      ;
                                                      if ((-1 < (int)*plVar13) &&
                                                         (*plVar13 = *plVar13 + -1, *plVar13 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar13);
                                                        (**(code **)(lVar11 + 0x30))(plVar13);
                                                      }
                                                      plVar7 = local_1a0;
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        if ((-1 < (int)*plVar10) &&
                                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar10);
                                                          (**(code **)(lVar11 + 0x30))(plVar10);
                                                        }
                                                        uVar18 = 0x336;
                                                        plVar7 = local_1a0;
                                                        pcVar8 = local_1a8;
                                                        plVar9 = local_198;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        uVar18 = 0x336;
                                                        *(undefined4 *)(local_1a0 + 5) = 0x336;
                                                        plVar12 = (longlong *)
                                                                  FUN_180376770(param_1,plVar10,
                                                                                plVar9);
                                                        if ((-1 < (int)*plVar10) &&
                                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar10);
                                                          (**(code **)(lVar11 + 0x30))(plVar10);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar11 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar11 + 0x30))(plVar9);
                                                        }
                                                        if (plVar12 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1b8);
                                                          pcVar8 = local_1a8;
                                                          plVar9 = local_198;
                                                          plVar12 = local_1b0;
                                                        }
                                                        else {
                                                          if ((-1 < (int)*plVar12) &&
                                                             (*plVar12 = *plVar12 + -1,
                                                             *plVar12 == 0)) {
                                                            lVar11 = FUN_18000cc40(plVar12);
                                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                                          }
                                                          plVar9 = local_198;
                                                          plVar12 = (longlong *)
                                                                    FUN_18037f0c0(param_1,local_198,
                                                                                  DAT_1803b4a18);
                                                          if (plVar12 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1b8);
                                                            uVar18 = 0x337;
                                                            plVar7 = local_1a0;
                                                            pcVar8 = local_1a8;
                                                            plVar12 = local_1b0;
                                                          }
                                                          else {
                                                            plVar9 = (longlong *)
                                                                     FUN_180374a40(param_1,4);
                                                            if (*(int *)local_1a8 + 1 != 0) {
                                                              *(int *)local_1a8 =
                                                                   *(int *)local_1a8 + 1;
                                                            }
                                                            FUN_18000cc80(plVar9,0,local_1a8);
                                                            if (*(int *)local_d8 + 1 != 0) {
                                                              *(int *)local_168 =
                                                                   *(int *)local_d8 + 1;
                                                            }
                                                            FUN_18000cc80(plVar9,1,local_168);
                                                            iVar5 = *(int *)local_1b0[2] + 1;
                                                            if (iVar5 != 0) {
                                                              *(int *)local_1b0[2] = iVar5;
                                                            }
                                                            FUN_18000cc80(plVar9,2);
                                                            if (*(int *)local_140 + 1 != 0) {
                                                              *(int *)local_140 =
                                                                   *(int *)local_140 + 1;
                                                            }
                                                            FUN_18000cc80(plVar9,3,local_140);
                                                            plVar7 = (longlong *)
                                                                     FUN_18036fb80(param_1,
                                                  DAT_1803b4b50);
                                                  uVar18 = 0x337;
                                                  *(undefined4 *)(local_1a0 + 5) = 0x337;
                                                  plVar10 = (longlong *)
                                                            FUN_180001790(param_1,plVar12,plVar9,
                                                                          plVar7);
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar11 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar11 + 0x30))(plVar7);
                                                  }
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                    }
                                                    plVar9 = local_198;
                                                    plVar10 = local_1a0;
                                                    uVar18 = 0x338;
                                                    *(undefined4 *)(local_1a0 + 5) = 0x338;
                                                    plVar12 = (longlong *)
                                                              FUN_18037e820(param_1,local_198,
                                                                            DAT_1803b4060,
                                                                            DAT_1803b41f0 + 0x18);
                                                    if (plVar12 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar12) &&
                                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar12);
                                                        (**(code **)(lVar11 + 0x30))(plVar12);
                                                      }
                                                      uVar18 = 0x339;
                                                      *(undefined4 *)(plVar10 + 5) = 0x339;
                                                      plVar12 = (longlong *)
                                                                FUN_18037e820(param_1,plVar9,
                                                                              DAT_1803b4060,
                                                                              DAT_1803b4200 + 0x18);
                                                      if (plVar12 != (longlong *)0x0) {
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar11 + 0x30))(plVar12);
                                                        }
                                                        uVar18 = 0x33a;
                                                        *(undefined4 *)(plVar10 + 5) = 0x33a;
                                                        plVar12 = (longlong *)
                                                                  FUN_18037e820(param_1,plVar9,
                                                                                DAT_1803b4060,
                                                                                DAT_1803b4b58 + 0x18
                                                                               );
                                                        if (plVar12 != (longlong *)0x0) {
                                                          if ((-1 < (int)*plVar12) &&
                                                             (*plVar12 = *plVar12 + -1,
                                                             *plVar12 == 0)) {
                                                            lVar11 = FUN_18000cc40(plVar12);
                                                            (**(code **)(lVar11 + 0x30))(plVar12);
                                                          }
                                                          uVar18 = 0x33c;
                                                          *(undefined4 *)(plVar10 + 5) = 0x33c;
                                                          plVar12 = (longlong *)
                                                                    FUN_18037e980(param_1,plVar9,
                                                                                  DAT_1803b4060,
                                                                                  DAT_1803b4b60 +
                                                                                  0x18);
                                                          if (plVar12 != (longlong *)0x0) {
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            plVar12 = (longlong *)
                                                                      FUN_18037f0c0(param_1,plVar9,
                                                                                    DAT_1803b3f10);
                                                            plVar7 = plVar10;
                                                            if (plVar12 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1b8);
                                                              uVar18 = 0x33d;
                                                              pcVar8 = local_1a8;
                                                              plVar12 = local_1b0;
                                                            }
                                                            else {
                                                              local_60 = local_1a8;
                                                              uVar18 = 0x33d;
                                                              *(undefined4 *)(plVar10 + 5) = 0x33d;
                                                              local_58 = local_110;
                                                              plVar9 = (longlong *)
                                                                       FUN_18037e820(param_1,plVar12
                                                                                     ,DAT_1803b41d8,
                                                                                     &local_60);
                                                              if ((-1 < (int)*plVar12) &&
                                                                 (*plVar12 = *plVar12 + -1,
                                                                 *plVar12 == 0)) {
                                                                lVar11 = FUN_18000cc40(plVar12);
                                                                (**(code **)(lVar11 + 0x30))
                                                                          (plVar12);
                                                              }
                                                              if (plVar9 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_1b8);
                                                                pcVar8 = local_1a8;
                                                                plVar9 = local_198;
                                                                plVar12 = local_1b0;
                                                              }
                                                              else {
                                                                if ((-1 < (int)*plVar9) &&
                                                                   (*plVar9 = *plVar9 + -1,
                                                                   *plVar9 == 0)) {
                                                                  lVar11 = FUN_18000cc40(plVar9);
                                                                  (**(code **)(lVar11 + 0x30))
                                                                            (plVar9);
                                                                }
                                                                lVar11 = FUN_1800e1020(param_1);
                                                                if (lVar11 == 0) {
                                                                  FUN_180373340(param_1,&local_1b8,
                                                                                DAT_1803b45d0);
                                                                  uVar18 = 0x33e;
                                                                  pcVar8 = local_1a8;
                                                                  plVar9 = local_198;
                                                                  plVar12 = local_1b0;
                                                                }
                                                                else {
                                                                  uVar18 = 0x33e;
                                                                  *(undefined4 *)(plVar10 + 5) =
                                                                       0x33e;
                                                                  plVar9 = (longlong *)
                                                                           FUN_180376210(param_1,
                                                  lVar11);
                                                  pcVar16 = local_168;
                                                  if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    cVar4 = FUN_18037f3a0(param_1,local_168,
                                                                          DAT_1803b45d8,plVar9);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar11 + 0x30))(plVar9);
                                                    }
                                                    plVar9 = local_198;
                                                    if (cVar4 == '\0') {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                    }
                                                    else {
                                                      plVar12 = (longlong *)
                                                                FUN_18037f0c0(param_1,local_198,
                                                                              DAT_1803b3f10);
                                                      if (plVar12 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x33f;
                                                        pcVar8 = local_1a8;
                                                        plVar12 = local_1b0;
                                                      }
                                                      else {
                                                        uVar18 = 0x33f;
                                                        *(undefined4 *)(plVar10 + 5) = 0x33f;
                                                        plVar9 = (longlong *)
                                                                 FUN_18037e6c0(param_1,plVar12,
                                                                               DAT_1803b4070,pcVar16
                                                                              );
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar11 + 0x30))(plVar12);
                                                        }
                                                        if (plVar9 != (longlong *)0x0) {
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar11 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar11 + 0x30))(plVar9);
                                                          }
                                                          plVar9 = local_198;
                                                          uVar18 = 0x340;
                                                          *(undefined4 *)(plVar10 + 5) = 0x340;
                                                          plVar12 = (longlong *)
                                                                    FUN_18037e820(param_1,local_198,
                                                                                  DAT_1803b4060,
                                                                                  DAT_1803b45e0 +
                                                                                  0x18);
                                                          if (plVar12 != (longlong *)0x0) {
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            plVar12 = local_1b0;
                                                            uVar18 = 0x341;
                                                            local_60 = local_1a8;
                                                            local_58 = local_168;
                                                            local_48 = (longlong *)local_1b0[2];
                                                            local_50 = local_140;
                                                            *(undefined4 *)(plVar10 + 5) = 0x341;
                                                            plVar13 = (longlong *)
                                                                      FUN_18037eae0(param_1,plVar9,
                                                                                    DAT_1803b45e8,
                                                                                    &local_60);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar11 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar11 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              plVar13 = (longlong *)
                                                                        FUN_18037f0c0(param_1,plVar9
                                                                                      ,DAT_1803b3fa8
                                                                                     );
                                                              pcVar3 = local_1a8;
                                                              pcVar2 = DAT_1803b4718;
                                                              pcVar16 = _Py_TrueStruct_exref;
                                                              if (plVar13 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_1b8);
                                                                uVar18 = 0x342;
                                                                pcVar8 = local_1a8;
                                                                plVar12 = local_1b0;
                                                              }
                                                              else {
                                                                local_48 = (longlong *)plVar12[2];
                                                                uVar18 = 0x342;
                                                                *(undefined4 *)(plVar10 + 5) = 0x342
                                                                ;
                                                                local_60 = pcVar2;
                                                                local_58 = pcVar16;
                                                                local_258 = (longlong *)
                                                                            DAT_1803b4b68;
                                                                local_50 = local_1a8;
                                                                plVar9 = (longlong *)
                                                                         FUN_18037d280(param_1,
                                                  plVar13,DAT_1803b4208,&local_60);
                                                  if ((-1 < (int)*plVar13) &&
                                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar13);
                                                    (**(code **)(lVar11 + 0x30))(plVar13);
                                                  }
                                                  if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    plVar12 = local_1b0;
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar11 + 0x30))(plVar9);
                                                    }
                                                    plVar12 = local_1b0;
                                                    uVar18 = DAT_1803b45f0;
                                                    lVar11 = local_1b0[2];
                                                    plVar10 = (longlong *)_PyDict_NewPresized(6);
                                                    PyDict_SetItem(plVar10,uVar18,lVar11);
                                                    PyDict_SetItem(plVar10,DAT_1803b4080,local_158);
                                                    uVar18 = DAT_1803b45f8;
                                                    plVar9 = (longlong *)
                                                             FUN_18037f0c0(param_1,local_150,
                                                                           DAT_1803b45f8);
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      uVar18 = 0x346;
                                                    }
                                                    else {
                                                      PyDict_SetItem(plVar10,uVar18,plVar9);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar11 + 0x30))(plVar9);
                                                      }
                                                      uVar18 = DAT_1803b4ae0;
                                                      plVar9 = (longlong *)
                                                               FUN_18037f0c0(param_1,local_168,
                                                                             DAT_1803b4080);
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x347;
                                                      }
                                                      else {
                                                        PyDict_SetItem(plVar10,uVar18,plVar9);
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar11 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar11 + 0x30))(plVar9);
                                                        }
                                                        plVar9 = local_108;
                                                        uVar18 = DAT_1803b43d8;
                                                        iVar5 = FUN_180001510(local_108);
                                                        if (iVar5 != -1) {
                                                          plVar7 = DAT_1803b4728;
                                                          if (iVar5 == 1) {
                                                            plVar7 = plVar9;
                                                          }
                                                          plVar9 = (longlong *)FUN_1803749a0(plVar7)
                                                          ;
                                                          if (plVar9 != (longlong *)0x0) {
                                                            PyDict_SetItem(plVar10,uVar18,plVar9);
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar11 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            PyDict_SetItem(plVar10,DAT_1803b4b70,
                                                                           plVar12[2]);
                                                            FUN_180001d80(param_1);
                                                            if ((-1 < *(int *)local_d8) &&
                                                               (*(longlong *)local_d8 =
                                                                     *(longlong *)local_d8 + -1,
                                                               *(longlong *)local_d8 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_168);
                                                              (**(code **)(lVar11 + 0x30))
                                                                        (local_168);
                                                            }
                                                            if ((-1 < *(int *)local_e8) &&
                                                               (*(longlong *)local_e8 =
                                                                     *(longlong *)local_e8 + -1,
                                                               *(longlong *)local_e8 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_110);
                                                              (**(code **)(lVar11 + 0x30))
                                                                        (local_110);
                                                            }
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar11 + 0x30))(plVar12);
                                                            }
                                                            if ((-1 < (int)*local_120) &&
                                                               (*local_120 = *local_120 + -1,
                                                               *local_120 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_138);
                                                              (**(code **)(lVar11 + 0x30))
                                                                        (local_138);
                                                            }
                                                            if ((-1 < (int)*local_128) &&
                                                               (*local_128 = *local_128 + -1,
                                                               *local_128 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_150);
                                                              (**(code **)(lVar11 + 0x30))
                                                                        (local_150);
                                                            }
                                                            pcVar16 = local_130;
                                                            if ((-1 < *(int *)local_130) &&
                                                               (*(longlong *)local_130 =
                                                                     *(longlong *)local_130 + -1,
                                                               *(longlong *)local_130 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_130);
                                                              (**(code **)(lVar11 + 0x30))(pcVar16);
                                                            }
                                                            plVar9 = local_100;
                                                            if (((local_100 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_100)) &&
                                                               (*local_100 = *local_100 + -1,
                                                               *local_100 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_100);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_f8;
                                                            if (((local_f8 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_f8)) &&
                                                               (*local_f8 = *local_f8 + -1,
                                                               *local_f8 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_f8);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_d0;
                                                            if ((-1 < (int)*local_d0) &&
                                                               (*local_d0 = *local_d0 + -1,
                                                               *local_d0 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_d0);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_188;
                                                            if ((-1 < (int)*local_188) &&
                                                               (*local_188 = *local_188 + -1,
                                                               *local_188 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_188);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_170;
                                                            if ((-1 < (int)*local_170) &&
                                                               (*local_170 = *local_170 + -1,
                                                               *local_170 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_170);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_158;
                                                            if ((-1 < (int)*local_158) &&
                                                               (*local_158 = *local_158 + -1,
                                                               *local_158 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_158);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_148;
                                                            if ((-1 < (int)*local_148) &&
                                                               (*local_148 = *local_148 + -1,
                                                               *local_148 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_148);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            if ((-1 < *(int *)pcVar8) &&
                                                               (*(longlong *)pcVar8 =
                                                                     *(longlong *)pcVar8 + -1,
                                                               *(longlong *)pcVar8 == 0)) {
                                                              lVar11 = FUN_18000cc40(pcVar8);
                                                              (**(code **)(lVar11 + 0x30))(pcVar8);
                                                            }
                                                            pcVar8 = local_140;
                                                            if ((-1 < *(int *)local_140) &&
                                                               (*(longlong *)local_140 =
                                                                     *(longlong *)local_140 + -1,
                                                               *(longlong *)local_140 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_140);
                                                              (**(code **)(lVar11 + 0x30))(pcVar8);
                                                            }
                                                            plVar9 = local_198;
                                                            if ((-1 < (int)*local_198) &&
                                                               (*local_198 = *local_198 + -1,
                                                               *local_198 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_198);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            if ((-1 < *(int *)pcVar3) &&
                                                               (*(longlong *)pcVar3 =
                                                                     *(longlong *)pcVar3 + -1,
                                                               *(longlong *)pcVar3 == 0)) {
                                                              lVar11 = FUN_18000cc40(pcVar3);
                                                              (**(code **)(lVar11 + 0x30))(pcVar3);
                                                            }
                                                            plVar9 = local_f0;
                                                            if ((-1 < (int)*local_f0) &&
                                                               (*local_f0 = *local_f0 + -1,
                                                               *local_f0 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_f0);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_108;
                                                            if ((-1 < (int)*local_108) &&
                                                               (*local_108 = *local_108 + -1,
                                                               *local_108 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_108);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            plVar9 = local_e0;
                                                            if ((-1 < (int)*local_e0) &&
                                                               (*local_e0 = *local_e0 + -1,
                                                               *local_e0 == 0)) {
                                                              lVar11 = FUN_18000cc40(local_e0);
                                                              (**(code **)(lVar11 + 0x30))(plVar9);
                                                            }
                                                            pcVar8 = local_118;
                                                            if (*(int *)local_118 < 0) {
                                                              return plVar10;
                                                            }
                                                            *(longlong *)local_118 =
                                                                 *(longlong *)local_118 + -1;
                                                            if (*(longlong *)local_118 != 0) {
                                                              return plVar10;
                                                            }
                                                            lVar11 = FUN_18000cc40(local_118);
                                                            (**(code **)(lVar11 + 0x30))(pcVar8);
                                                            return plVar10;
                                                          }
                                                        }
                                                        FUN_18000ce30(param_1,&local_1b8);
                                                        uVar18 = 0x348;
                                                      }
                                                    }
                                                    plVar7 = local_1a0;
                                                    pcVar8 = local_1a8;
                                                    plVar9 = local_198;
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                      plVar7 = local_1a0;
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                    }
                                                  }
                                                  }
                                                  goto LAB_1801006e3;
                                                  }
                                                  }
                                                  goto LAB_180100c7d;
                                                  }
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801006e3;
                                                  }
                                                  }
                                                  }
                                                  }
LAB_180100c7d:
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  plVar7 = plVar10;
                                                  pcVar8 = local_1a8;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801006e3;
                                                  }
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  plVar7 = local_1a0;
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  if (((int)*plVar13 < 0) ||
                                                     (*plVar13 = *plVar13 + -1, *plVar13 != 0))
                                                  goto LAB_1801006e3;
                                                  goto LAB_180101752;
                                                  }
                                                  plVar9 = (longlong *)
                                                           FUN_1800b4830(param_1,plVar9,
                                                                         DAT_1803ab250,0);
                                                  if (plVar9 == (longlong *)0x0) {
LAB_180101c03:
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                  }
                                                  else {
                                                    plVar10 = (longlong *)
                                                              FUN_18037f0c0(param_1,plVar9,
                                                                            DAT_1803b49c0);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar11 + 0x30))(plVar9);
                                                    }
                                                    if (plVar10 == (longlong *)0x0)
                                                    goto LAB_180101c03;
                                                    uVar18 = 0x32b;
                                                    *(undefined4 *)(plVar7 + 5) = 0x32b;
                                                    plVar17 = (longlong *)
                                                              FUN_1803770d0(param_1,plVar10,
                                                                            DAT_1803b49c8);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar11 + 0x30))(plVar10);
                                                    }
                                                    if (plVar17 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1b8);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                      goto LAB_1801006e3;
                                                    }
                                                    iVar6 = FUN_180001510(plVar17);
                                                    if (iVar6 != -1) {
                                                      plVar9 = local_170;
                                                      if ((-1 < (int)*plVar17) &&
                                                         (*plVar17 = *plVar17 + -1, *plVar17 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar17);
                                                        (**(code **)(lVar11 + 0x30))(plVar17);
                                                        plVar9 = local_170;
                                                      }
                                                      goto LAB_180101d24;
                                                    }
                                                    FUN_18000ce30(param_1,&local_1b8);
                                                    if ((-1 < (int)*plVar17) &&
                                                       (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar17);
                                                      (**(code **)(lVar11 + 0x30))(plVar17);
                                                      pcVar8 = local_1a8;
                                                      plVar9 = local_198;
                                                      plVar12 = local_1b0;
                                                      goto LAB_1801006e3;
                                                    }
                                                  }
                                                  uVar18 = 0x32b;
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801006e3;
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_1b8);
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar11 + 0x30))(plVar12);
                                                  }
                                                  uVar18 = 0x324;
                                                  pcVar8 = local_1a8;
                                                  plVar9 = local_198;
                                                  plVar12 = local_1b0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                }
                                                goto LAB_1801006e3;
                                              }
                                            }
                                          }
                                          FUN_18000ce30(param_1,&local_1b8);
                                          uVar18 = 0x315;
                                          pcVar8 = local_1a8;
                                          plVar12 = local_1b0;
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                              goto LAB_1801006e3;
                            }
                          }
                          uVar18 = 0x309;
                        }
                      }
LAB_1801006c7:
                      local_1b8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar7 = local_1a0;
                      pcVar8 = local_1a8;
                      plVar9 = local_198;
                      plVar12 = local_1b0;
                    }
                  }
                  goto LAB_1801006e3;
                }
              }
              else {
                plVar10 = (longlong *)FUN_1800b4830(param_1,pcVar8,DAT_1803ab260,0xffffffffffffffff)
                ;
                if (plVar10 != (longlong *)0x0) {
                  plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803b4a48);
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                  if (plVar7 != (longlong *)0x0) goto LAB_18010045f;
                }
              }
            }
            local_1b8 = *(longlong *)(param_1 + 0x60);
            uVar18 = 0x305;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar7 = local_1a0;
            pcVar8 = local_1a8;
            plVar12 = local_1b0;
          }
        }
      }
    }
  }
LAB_1801006e3:
  piVar1 = *(int **)(local_1b8 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar11 = FUN_18038e2a0(plVar7,uVar18);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar7) goto LAB_1801033e2;
    lVar11 = FUN_18038e2a0(plVar7,uVar18);
    *(int **)(lVar11 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_1b8,lVar11);
LAB_1801033e2:
  pcVar16 = local_118;
  local_1c8 = local_140;
  local_1d0 = local_b8;
  local_1d8 = local_148;
  local_1e0 = local_158;
  local_1e8 = local_c0;
  local_1f0 = local_a8;
  local_1f8 = local_b0;
  local_200 = local_f8;
  local_208 = local_100;
  local_210 = local_130;
  local_218 = local_150;
  local_220 = local_138;
  local_230 = local_110;
  local_238 = local_168;
  local_240 = local_118;
  local_248 = local_e0;
  local_250 = (longlong **)local_108;
  local_258 = local_f0;
  local_228 = plVar12;
  FUN_18036d0a0(plVar7,"oooooooocoooooooooooo",plVar9,pcVar8);
  plVar10 = DAT_1803b3b30;
  if (plVar7 == DAT_1803b3b30) {
    if ((-1 < (int)*DAT_1803b3b30) && (*DAT_1803b3b30 = *DAT_1803b3b30 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803b3b30 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar11 = local_1b8;
  local_1b8 = 0;
  if (((local_168 != (code *)0x0) && (-1 < *(int *)local_168)) &&
     (*(longlong *)local_168 = *(longlong *)local_168 + -1, *(longlong *)local_168 == 0)) {
    (**(code **)(*(longlong *)(local_168 + 8) + 0x30))();
  }
  if (((local_110 != (code *)0x0) && (-1 < *(int *)local_110)) &&
     (*(longlong *)local_110 = *(longlong *)local_110 + -1, *(longlong *)local_110 == 0)) {
    (**(code **)(*(longlong *)(local_110 + 8) + 0x30))();
  }
  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))(plVar12);
  }
  if (((local_138 != (longlong *)0x0) && (-1 < (int)*local_138)) &&
     (*local_138 = *local_138 + -1, *local_138 == 0)) {
    (**(code **)(local_138[1] + 0x30))();
  }
  if (((local_150 != (longlong *)0x0) && (-1 < (int)*local_150)) &&
     (*local_150 = *local_150 + -1, *local_150 == 0)) {
    (**(code **)(local_150[1] + 0x30))();
  }
  if (((local_130 != (code *)0x0) && (-1 < *(int *)local_130)) &&
     (*(longlong *)local_130 = *(longlong *)local_130 + -1, *(longlong *)local_130 == 0)) {
    (**(code **)(*(longlong *)(local_130 + 8) + 0x30))();
  }
  if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
     (*local_100 = *local_100 + -1, *local_100 == 0)) {
    (**(code **)(local_100[1] + 0x30))();
  }
  if (((local_f8 != (longlong *)0x0) && (-1 < (int)*local_f8)) &&
     (*local_f8 = *local_f8 + -1, *local_f8 == 0)) {
    (**(code **)(local_f8[1] + 0x30))();
  }
  if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
     (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
    (**(code **)(local_b0[1] + 0x30))();
  }
  if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
     (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
    (**(code **)(local_a8[1] + 0x30))();
  }
  if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
     (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
    (**(code **)(local_c0[1] + 0x30))();
  }
  if (((local_158 != (longlong *)0x0) && (-1 < (int)*local_158)) &&
     (*local_158 = *local_158 + -1, *local_158 == 0)) {
    (**(code **)(local_158[1] + 0x30))();
  }
  if (((local_148 != (longlong *)0x0) && (-1 < (int)*local_148)) &&
     (*local_148 = *local_148 + -1, *local_148 == 0)) {
    (**(code **)(local_148[1] + 0x30))();
  }
  if (((local_b8 != (code *)0x0) && (-1 < *(int *)local_b8)) &&
     (*(longlong *)local_b8 = *(longlong *)local_b8 + -1, *(longlong *)local_b8 == 0)) {
    (**(code **)(*(longlong *)(local_b8 + 8) + 0x30))();
  }
  if (((local_140 != (code *)0x0) && (-1 < *(int *)local_140)) &&
     (*(longlong *)local_140 = *(longlong *)local_140 + -1, *(longlong *)local_140 == 0)) {
    (**(code **)(*(longlong *)(local_140 + 8) + 0x30))();
  }
  local_1b8 = lVar11;
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < *(int *)pcVar8) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
  if ((-1 < (int)*local_f0) && (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
    (**(code **)(local_f0[1] + 0x30))();
  }
  if ((-1 < (int)*local_108) && (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  if ((-1 < (int)*local_e0) && (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    (**(code **)(local_e0[1] + 0x30))();
  }
  if ((-1 < *(int *)pcVar16) &&
     (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
  }
  FUN_18000ce40(param_1,&local_1b8);
  return (longlong *)0x0;
}

